package templates

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/flow-hydraulics/flow-wallet-api/templates/template_strings"
	"github.com/onflow/flow-go-sdk"
	log "github.com/sirupsen/logrus"
)

type Token struct {
	ID                 uint64    `json:"id,omitempty"`
	Name               string    `json:"name" gorm:"uniqueIndex;not null"` // Declaration name
	NameLowerCase      string    `json:"nameLowerCase,omitempty"`          // (deprecated) For generic fungible token transaction templates
	ReceiverPublicPath string    `json:"receiverPublicPath,omitempty"`
	BalancePublicPath  string    `json:"balancePublicPath,omitempty"`
	VaultStoragePath   string    `json:"vaultStoragePath,omitempty"`
	Address            string    `json:"address" gorm:"not null"`
	Setup              string    `json:"setup,omitempty"`    // Setup cadence code
	Transfer           string    `json:"transfer,omitempty"` // Transfer cadence code
	Balance            string    `json:"balance,omitempty"`  // Balance cadence code
	Type               TokenType `json:"type"`
}

// BasicToken is a simplifed representation of a Token used in listings
type BasicToken struct {
	ID      uint64    `json:"id,omitempty"`
	Name    string    `json:"name"`
	Address string    `json:"address"`
	Type    TokenType `json:"type"`
}

type chainReplacers map[flow.ChainID]*strings.Replacer
type knownAddresses map[flow.ChainID]string
type templateVariables map[string]knownAddresses

var chains []flow.ChainID = []flow.ChainID{
	flow.Emulator,
	flow.Testnet,
	flow.Mainnet,
}

var knownAddressesReplacers chainReplacers

func makeReplacers(t templateVariables) chainReplacers {
	r := make(chainReplacers, len(chains))
	for _, c := range chains {
		vv := make([]string, len(t)*2)
		for varname, addressInChain := range t {
			vv = append(vv, varname, addressInChain[c])
		}
		r[c] = strings.NewReplacer(vv...)
	}
	return r
}

func (token Token) BasicToken() BasicToken {
	return BasicToken{
		ID:      token.ID,
		Name:    token.Name,
		Address: token.Address,
		Type:    token.Type,
	}
}

func TokenCode(chainId flow.ChainID, token *Token, tmplStr string) (string, error) {

	// Regex that matches all references to cadence source files
	// For example:
	// - "../../contracts/Source.cdc"
	// - "./Source.cdc"
	// - "Source.cdc"
	matchCadenceFiles := regexp.MustCompile(`"(.*?)(\w+\.cdc)"`)

	// Replace all above matches with just the filename, without quotes
	replaceCadenceFiles := "$2"

	// Replaces all TokenName.cdc's with TOKEN_ADDRESS
	sourceFileReplacer := strings.NewReplacer(
		fmt.Sprintf("%s.cdc", token.Name), "TOKEN_ADDRESS",
	)

	tokenVault, tokenReceiver, tokenBalance, err := GetTokenPaths(token)
	if err != nil && (strings.Contains(tmplStr, "TOKEN_VAULT") ||
		strings.Contains(tmplStr, "TOKEN_RECEIVER") ||
		strings.Contains(tmplStr, "TOKEN_BALANCE")) {

		return "", err
	}

	templateReplacer := strings.NewReplacer(
		"TOKEN_DECLARATION_NAME", token.Name,
		"TOKEN_ADDRESS", token.Address,
		"TOKEN_VAULT", tokenVault,
		"TOKEN_RECEIVER", tokenReceiver,
		"TOKEN_BALANCE", tokenBalance,
	)

	knownAddressesReplacer := knownAddressesReplacers[chainId]

	code := tmplStr

	// Ordering matters here
	code = matchCadenceFiles.ReplaceAllString(code, replaceCadenceFiles)
	code = sourceFileReplacer.Replace(code)
	code = templateReplacer.Replace(code)
	code = knownAddressesReplacer.Replace(code)

	return code, nil
}

func GetTokenPaths(
	token *Token,
) (
	vaultPath string,
	receiverPath string,
	balancePath string,
	err error,
) {

	if token.VaultStoragePath != "" &&
		token.ReceiverPublicPath != "" &&
		token.BalancePublicPath != "" {

		// All three paths are set explicitly.

		vaultPath = token.VaultStoragePath
		receiverPath = token.ReceiverPublicPath
		balancePath = token.BalancePublicPath

	} else if token.NameLowerCase != "" &&
		token.VaultStoragePath == "" &&
		token.ReceiverPublicPath == "" &&
		token.BalancePublicPath == "" {

		// Deprecated config format: None of the paths are set.
		// Use token.NameLowerCase to generate paths.

		if token.Name != "FlowToken" && token.Name != "FUSD" {
			log.Warnf("%s token is using deprecated config format", token.Name)
		}

		vaultPath = fmt.Sprintf("/storage/%sVault", token.NameLowerCase)
		receiverPath = fmt.Sprintf("/public/%sReceiver", token.NameLowerCase)
		balancePath = fmt.Sprintf("/public/%sBalance", token.NameLowerCase)

	} else {

		// Configuration error in paths

		err = fmt.Errorf("invalid path configuration for token %s", token.Name)
		return
	}

	return
}

func FungibleTransferCode(chainId flow.ChainID, token *Token) (string, error) {
	return TokenCode(chainId, token, template_strings.GenericFungibleTransfer)
}

func FungibleSetupCode(chainId flow.ChainID, token *Token) (string, error) {
	return TokenCode(chainId, token, template_strings.GenericFungibleSetup)
}

func FungibleBalanceCode(chainId flow.ChainID, token *Token) (string, error) {
	return TokenCode(chainId, token, template_strings.GenericFungibleBalance)
}

func InitFungibleTokenVaultsCode(chainId flow.ChainID, tokens []template_strings.FungibleTokenInfo) (string, error) {
	return template_strings.AddFungibleTokenVaultBatchTransaction(template_strings.BatchedFungibleOpsInfo{
		FungibleTokenContractAddress: KnownAddresses["FungibleToken.cdc"][chainId],
		Tokens:                       tokens,
	})
}

func CreateAccountAndInitFungibleTokenVaultsCode(chainId flow.ChainID, tokens []template_strings.FungibleTokenInfo) (string, error) {
	return template_strings.CreateAccountAndSetupTransaction(template_strings.BatchedFungibleOpsInfo{
		FungibleTokenContractAddress: KnownAddresses["FungibleToken.cdc"][chainId],
		Tokens:                       tokens,
	})
}
