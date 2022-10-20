package accounts

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/flow-hydraulics/flow-wallet-api/configs"
	"github.com/flow-hydraulics/flow-wallet-api/datastore"
	"github.com/flow-hydraulics/flow-wallet-api/flow_helpers"
	"github.com/flow-hydraulics/flow-wallet-api/jobs"
	"github.com/flow-hydraulics/flow-wallet-api/keys"
	"github.com/flow-hydraulics/flow-wallet-api/templates"
	"github.com/flow-hydraulics/flow-wallet-api/templates/template_strings"
	"github.com/flow-hydraulics/flow-wallet-api/transactions"
	"github.com/onflow/cadence"
	jsoncdc "github.com/onflow/cadence/encoding/json"
	"github.com/onflow/flow-go-sdk"
	flow_crypto "github.com/onflow/flow-go-sdk/crypto"
	flow_templates "github.com/onflow/flow-go-sdk/templates"
	log "github.com/sirupsen/logrus"
	"go.uber.org/ratelimit"
)

const maxGasLimit = 9999

type Service interface {
	List(limit, offset int) (result []Account, err error)
	Create(ctx context.Context, sync bool) (*jobs.Job, *Account, error)
	AddNonCustodialAccount(address string) (*Account, error)
	DeleteNonCustodialAccount(address string) error
	SyncAccountKeyCount(ctx context.Context, address flow.Address) (*jobs.Job, error)
	Details(address string) (Account, error)
	InitAdminAccount(ctx context.Context) error
}

// ServiceImpl defines the API for account management.
type ServiceImpl struct {
	cfg           *configs.Config
	store         Store
	km            keys.Manager
	fc            flow_helpers.FlowClient
	wp            jobs.WorkerPool
	txs           transactions.Service
	temps         templates.Service
	txRateLimiter ratelimit.Limiter
}

// NewService initiates a new account service.
func NewService(
	cfg *configs.Config,
	store Store,
	km keys.Manager,
	fc flow_helpers.FlowClient,
	wp jobs.WorkerPool,
	txs transactions.Service,
	temps templates.Service,
	opts ...ServiceOption,
) Service {
	var defaultTxRatelimiter = ratelimit.NewUnlimited()

	// TODO(latenssi): safeguard against nil config?
	svc := &ServiceImpl{cfg, store, km, fc, wp, txs, temps, defaultTxRatelimiter}

	for _, opt := range opts {
		opt(svc)
	}

	if wp == nil {
		panic("workerpool nil")
	}

	// Register asynchronous job executors
	wp.RegisterExecutor(AccountCreateJobType, svc.executeAccountCreateJob)
	wp.RegisterExecutor(SyncAccountKeyCountJobType, svc.executeSyncAccountKeyCountJob)

	return svc
}

// List returns all accounts in the datastore.
func (s *ServiceImpl) List(limit, offset int) (result []Account, err error) {
	o := datastore.ParseListOptions(limit, offset)
	return s.store.Accounts(o)
}

// Create calls account.New to generate a new account.
// It receives a new account with a corresponding private key or resource ID
// and stores both in datastore.
// It returns a job, the new account and a possible error.
func (s *ServiceImpl) Create(ctx context.Context, sync bool) (*jobs.Job, *Account, error) {
	log.WithFields(log.Fields{"sync": sync}).Trace("Create account")

	if !sync {
		job, err := s.wp.CreateJob(AccountCreateJobType, "")
		if err != nil {
			return nil, nil, err
		}

		err = s.wp.Schedule(job)
		if err != nil {
			return nil, nil, err
		}

		return job, nil, err
	}

	account, _, err := s.createAccount(ctx)
	if err != nil {
		return nil, nil, err
	}

	return nil, account, nil
}

func (s *ServiceImpl) AddNonCustodialAccount(address string) (*Account, error) {
	log.WithFields(log.Fields{"address": address}).Trace("Add non-custodial account")

	a := &Account{
		Address: flow_helpers.HexString(address),
		Type:    AccountTypeNonCustodial,
	}

	err := s.store.InsertAccount(a)
	if err != nil {
		return nil, err
	}

	return a, nil
}

func (s *ServiceImpl) DeleteNonCustodialAccount(address string) error {
	log.WithFields(log.Fields{"address": address}).Trace("Delete non-custodial account")

	a, err := s.store.Account(flow_helpers.HexString(address))
	if err != nil {
		if strings.Contains(err.Error(), "record not found") {
			// Account already gone. All good.
			return nil
		}

		return err
	}

	if a.Type != AccountTypeNonCustodial {
		return fmt.Errorf("only non-custodial accounts supported")
	}

	return s.store.HardDeleteAccount(&a)
}

// Details returns a specific account, does not include private keys
func (s *ServiceImpl) Details(address string) (Account, error) {
	log.WithFields(log.Fields{"address": address}).Trace("Account details")

	// Check if the input is a valid address
	address, err := flow_helpers.ValidateAddress(address, s.cfg.ChainID)
	if err != nil {
		return Account{}, err
	}

	account, err := s.store.Account(address)
	if err != nil {
		return Account{}, err
	}

	// Strip the private keys
	for i := range account.Keys {
		account.Keys[i].Value = make([]byte, 0)
	}

	return account, nil
}

// SyncKeyCount syncs number of keys for given account
func (s *ServiceImpl) SyncAccountKeyCount(ctx context.Context, address flow.Address) (*jobs.Job, error) {
	// Validate address, they might be legit addresses but for the wrong chain
	if !address.IsValid(s.cfg.ChainID) {
		return nil, fmt.Errorf(`not a valid address for %s: "%s"`, s.cfg.ChainID, address)
	}

	// Prepare job attributes required for executing the job
	attrs := syncAccountKeyCountJobAttributes{Address: address, NumKeys: int(s.cfg.DefaultAccountKeyCount)}
	attrBytes, err := json.Marshal(attrs)
	if err != nil {
		return nil, err
	}

	// Create & schedule the "sync key count" job
	job, err := s.wp.CreateJob(SyncAccountKeyCountJobType, "", jobs.WithAttributes(attrBytes))
	if err != nil {
		return nil, err
	}
	err = s.wp.Schedule(job)
	if err != nil {
		return nil, err
	}

	return job, nil
}

// syncAccountKeyCount syncs the number of account keys with the given numKeys and
// returns the number of keys, transaction ID and error.
func (s *ServiceImpl) syncAccountKeyCount(ctx context.Context, address flow.Address, numKeys int) (int, string, error) {
	entry := log.WithFields(log.Fields{"address": address, "numKeys": numKeys, "function": "ServiceImpl.syncAccountKeyCount"})

	if numKeys < 1 {
		return 0, "", fmt.Errorf("invalid number of keys specified: %d, min. 1 expected", numKeys)
	}

	// Check on-chain keys
	flowAccount, err := s.fc.GetAccount(ctx, address)
	if err != nil {
		entry.WithFields(log.Fields{"err": err}).Error("failed to get Flow account")
		return 0, "", err
	}

	// Get stored account
	dbAccount, err := s.store.Account(flow_helpers.FormatAddress(address))
	if err != nil {
		entry.WithFields(log.Fields{"err": err}).Error("failed to get account from database")
		return 0, "", err
	}

	// Pick a source key that will be used to create the new keys & decode public key
	sourceKey := dbAccount.Keys[0] // NOTE: Only valid (not revoked) keys should be stored in the database
	sourceKeyPbkString := strings.TrimPrefix(sourceKey.PublicKey, "0x")
	sourcePbk, err := flow_crypto.DecodePublicKeyHex(flow_crypto.StringToSignatureAlgorithm(sourceKey.SignAlgo), sourceKeyPbkString)
	if err != nil {
		entry.WithFields(log.Fields{"err": err, "sourceKeyPbkString": sourceKeyPbkString}).Error("failed to decode public key for source key")
		return 0, "", err
	}
	entry.WithFields(log.Fields{"sourceKeyId": sourceKey.ID, "sourcePbk": sourcePbk}).Trace("source key selected")

	// Count valid keys, as some keys might be revoked, assuming dbAccount.Keys are clones (all have same public key)
	var validKeys []*flow.AccountKey
	for i := range flowAccount.Keys {
		key := flowAccount.Keys[i]
		if !key.Revoked && key.PublicKey.Equals(sourcePbk) {
			validKeys = append(validKeys, key)
		}
	}

	if len(validKeys) != len(dbAccount.Keys) {
		entry.WithFields(log.Fields{"onChain": len(validKeys), "database": len(dbAccount.Keys)}).Warn("on-chain vs. database key count mismatch")
	}

	entry.WithFields(log.Fields{"validKeys": validKeys}).Trace("filtered valid keys")

	// Add keys by cloning the source key
	if len(validKeys) < numKeys {

		cloneCount := numKeys - len(validKeys)
		code := template_strings.AddAccountKeysTransaction
		pbks := []cadence.Value{}

		entry.WithFields(log.Fields{"validKeys": len(validKeys), "numKeys": numKeys, "cloneCount": cloneCount}).Debug("going to add keys")

		// Sort keys by index
		sort.SliceStable(dbAccount.Keys, func(i, j int) bool {
			return dbAccount.Keys[i].Index < dbAccount.Keys[j].Index
		})

		// Push publickeys to args and prepare db update
		for i := 0; i < cloneCount; i++ {
			pbk, err := cadence.NewString(sourceKey.PublicKey[2:]) // TODO: use a helper function to trim "0x" prefix
			if err != nil {
				return 0, "", err
			}
			pbks = append(pbks, pbk)

			// Create cloned account key & update index
			cloned := keys.Storable{
				ID:             0, // Reset ID to create a new key to DB
				AccountAddress: sourceKey.AccountAddress,
				Index:          dbAccount.Keys[len(dbAccount.Keys)-1].Index + 1,
				Type:           sourceKey.Type,
				Value:          sourceKey.Value,
				PublicKey:      sourceKey.PublicKey,
				SignAlgo:       sourceKey.SignAlgo,
				HashAlgo:       sourceKey.HashAlgo,
			}

			dbAccount.Keys = append(dbAccount.Keys, cloned)
		}

		// Prepare transaction arguments
		x := cadence.NewArray(pbks)
		args := []transactions.Argument{x}

		entry.WithFields(log.Fields{"args": args}).Debug("args prepared")

		// NOTE: sync, so will wait for transaction to be sent & sealed
		_, tx, err := s.txs.Create(ctx, true, dbAccount.Address, code, args, transactions.General)
		if err != nil {
			entry.WithFields(log.Fields{"err": err}).Error("failed to create transaction")
			return 0, tx.TransactionId, err
		}

		// Update account in database
		// TODO: if update fails, should sync keys from chain later
		err = s.store.SaveAccount(&dbAccount)
		if err != nil {
			entry.WithFields(log.Fields{"err": err}).Error("failed to update account in database")
			return 0, tx.TransactionId, err
		}

		return len(dbAccount.Keys), tx.TransactionId, err
	} else if len(validKeys) > numKeys {
		entry.Debug("too many valid keys", len(validKeys), " vs. ", numKeys)
	} else {
		entry.Debug("correct number of keys")
		return numKeys, "", nil
	}

	return 0, "", nil
}

// createAccount creates a new account on the flow blockchain. It generates a
// fresh key pair and constructs a flow transaction to create the account with
// generated key. Admin account is used to pay for the transaction.
//
// Returns created account and the flow transaction ID of the account creation.
func (s *ServiceImpl) createAccount(ctx context.Context) (*Account, string, error) {
	account := &Account{Type: AccountTypeCustodial}

	// Important to ratelimit all the way up here so the keys and reference blocks
	// are "fresh" when the transaction is actually sent
	s.txRateLimiter.Take()

	payer, err := s.km.AdminAuthorizer(ctx)
	if err != nil {
		return nil, "", err
	}

	proposer, err := s.km.AdminProposalKey(ctx)
	if err != nil {
		return nil, "", err
	}

	// Get latest blocks blockID as reference blockID
	referenceBlockID, err := flow_helpers.LatestBlockId(ctx, s.fc)
	if err != nil {
		return nil, "", err
	}

	// Generate a new key pair
	accountKey, newPrivateKey, err := s.km.GenerateDefault(ctx)
	if err != nil {
		return nil, "", err
	}

	// Public keys for creating the account
	publicKeys := []*flow.AccountKey{}

	// Create copies based on the configured key count, changing just the index
	for i := 0; i < int(s.cfg.DefaultAccountKeyCount); i++ {
		clonedAccountKey := *accountKey
		clonedAccountKey.Index = i

		publicKeys = append(publicKeys, &clonedAccountKey)
	}

	var flowTx *flow.Transaction
	var initializedFungibleTokens []templates.Token
	if s.cfg.InitFungibleTokenVaultsOnAccountCreation {

		flowTx, initializedFungibleTokens, err = s.generateCreateAccountTransactionWithEnabledFungibleTokenVaults(
			publicKeys,
			payer.Address,
		)
		if err != nil {
			return nil, "", err
		}

	} else {

		flowTx, err = flow_templates.CreateAccount(
			publicKeys,
			nil,
			payer.Address,
		)
		if err != nil {
			return nil, "", err
		}

	}

	flowTx.
		SetReferenceBlockID(*referenceBlockID).
		SetProposalKey(proposer.Address, proposer.Key.Index, proposer.Key.SequenceNumber).
		SetPayer(payer.Address).
		SetGasLimit(maxGasLimit)

	// Check if we want to use a custom account create script
	if s.cfg.ScriptPathCreateAccount != "" {
		bytes, err := os.ReadFile(s.cfg.ScriptPathCreateAccount)
		if err != nil {
			return nil, "", err
		}
		// Overwrite the existing script
		flowTx.SetScript(bytes)
	}

	// Proposer signs the payload (unless proposer == payer).
	if !proposer.Equals(payer) {
		if err := flowTx.SignPayload(proposer.Address, proposer.Key.Index, proposer.Signer); err != nil {
			return nil, "", err
		}
	}

	// Payer signs the envelope
	if err := flowTx.SignEnvelope(payer.Address, payer.Key.Index, payer.Signer); err != nil {
		return nil, "", err
	}

	// Send and wait for the transaction to be sealed
	result, err := flow_helpers.SendAndWait(ctx, s.fc, *flowTx, s.cfg.TransactionTimeout)
	if err != nil {
		return nil, "", err
	}

	// Grab the new address from transaction events
	var newAddress flow.Address
	for _, event := range result.Events {
		if event.Type == flow.EventAccountCreated {
			accountCreatedEvent := flow.AccountCreatedEvent(event)
			newAddress = accountCreatedEvent.Address()
			break
		}
	}

	// Check that we actually got a new address
	if newAddress == flow.EmptyAddress {
		return nil, "", fmt.Errorf("something went wrong when waiting for address")
	}

	account.Address = flow_helpers.FormatAddress(newAddress)

	// Convert the key to storable form (encrypt it)
	encryptedAccountKey, err := s.km.Save(*newPrivateKey)
	if err != nil {
		return nil, "", err
	}
	encryptedAccountKey.PublicKey = accountKey.PublicKey.String()

	// Store account and key(s)
	// Looping through accountKeys to get the correct Index values
	storableKeys := []keys.Storable{}
	for _, pbk := range publicKeys {
		clonedEncryptedAccountKey := encryptedAccountKey
		clonedEncryptedAccountKey.Index = pbk.Index
		storableKeys = append(storableKeys, clonedEncryptedAccountKey)
	}

	account.Keys = storableKeys
	if err := s.store.InsertAccount(account); err != nil {
		return nil, "", err
	}

	AccountAdded.Trigger(AccountAddedPayload{
		Address:                   flow.HexToAddress(account.Address),
		InitializedFungibleTokens: initializedFungibleTokens,
	})

	log.
		WithFields(log.Fields{"address": account.Address, "initialized-fungible-tokens": initializedFungibleTokens}).
		Info("Account created")

	return account, flowTx.ID().String(), nil
}

// generateCreateAccountTransactionWithEnabledFungibleTokenVaults is a helper function that generates a templated
// account creation transaction that initializes all enabled fungible tokens.
func (s *ServiceImpl) generateCreateAccountTransactionWithEnabledFungibleTokenVaults(
	publicKeys []*flow.AccountKey,
	payerAddress flow.Address,
) (
	*flow.Transaction,
	[]templates.Token,
	error,
) {
	// Create custom cadence script to create account and init enabled fungible tokens vaults
	tokens, err := s.temps.ListTokensFull(templates.FT)
	if err != nil {
		return nil, []templates.Token{}, nil
	}

	var initializedTokens []templates.Token
	tokensInfo := []template_strings.FungibleTokenInfo{}
	for _, t := range tokens {
		if t.Name != "FlowToken" {
			tokensInfo = append(tokensInfo, templates.NewFungibleTokenInfo(t))
			initializedTokens = append(initializedTokens, t)
		}
	}

	txScript, err := templates.CreateAccountAndInitFungibleTokenVaultsCode(s.cfg.ChainID, tokensInfo)
	if err != nil {
		return nil, []templates.Token{}, err
	}

	// Encode public key list
	keyList := make([]cadence.Value, len(publicKeys))
	for i, key := range publicKeys {
		keyList[i], err = flow_templates.AccountKeyToCadenceCryptoKey(key)
		if err != nil {
			return nil, []templates.Token{}, err
		}
	}
	cadencePublicKeys := cadence.NewArray(keyList)

	flowTx := flow.NewTransaction().
		SetScript([]byte(txScript)).
		AddAuthorizer(payerAddress).
		AddRawArgument(jsoncdc.MustEncode(cadencePublicKeys))

	return flowTx, initializedTokens, nil
}
