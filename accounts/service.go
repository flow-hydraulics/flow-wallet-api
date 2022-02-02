package accounts

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/flow-hydraulics/flow-wallet-api/configs"
	"github.com/flow-hydraulics/flow-wallet-api/datastore"
	"github.com/flow-hydraulics/flow-wallet-api/flow_helpers"
	"github.com/flow-hydraulics/flow-wallet-api/jobs"
	"github.com/flow-hydraulics/flow-wallet-api/keys"
	"github.com/onflow/flow-go-sdk"
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
	txRateLimiter ratelimit.Limiter
}

// NewService initiates a new account service.
func NewService(
	cfg *configs.Config,
	store Store,
	km keys.Manager,
	fc flow_helpers.FlowClient,
	wp jobs.WorkerPool,
	opts ...ServiceOption,
) Service {
	var defaultTxRatelimiter = ratelimit.NewUnlimited()

	// TODO(latenssi): safeguard against nil config?
	svc := &ServiceImpl{cfg, store, km, fc, wp, defaultTxRatelimiter}

	for _, opt := range opts {
		opt(svc)
	}

	if wp == nil {
		panic("workerpool nil")
	}

	// Register asynchronous job executor.
	wp.RegisterExecutor(AccountCreateJobType, svc.executeAccountCreateJob)

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

	// Generate a new key pair
	accountKey, newPrivateKey, err := s.km.GenerateDefault(ctx)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Failed to generate key")
		return nil, "", err
	}

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

	// Public keys for creating the account
	publicKeys := []*flow.AccountKey{}

	// Create copies based on the configured key count, changing just the index
	for i := 0; i < int(s.cfg.DefaultAccountKeyCount); i++ {
		clonedAccountKey := *accountKey
		clonedAccountKey.Index = i

		publicKeys = append(publicKeys, &clonedAccountKey)
	}

	flowTx := flow_templates.CreateAccount(
		publicKeys,
		nil,
		payer.Address,
	)

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
		log.WithFields(log.Fields{"error": err}).Error("Failed to send account create tx")
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
		Address: flow.HexToAddress(account.Address),
	})

	log.WithFields(log.Fields{"address": account.Address}).Debug("Account created")

	return account, flowTx.ID().String(), nil
}
