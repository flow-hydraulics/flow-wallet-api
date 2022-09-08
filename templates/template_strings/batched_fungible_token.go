package template_strings

import (
	"bytes"
	"text/template"
)

type FungibleTokenInfo struct {
	ContractName       string
	Address            string
	VaultStoragePath   string
	ReceiverPublicPath string
	BalancePublicPath  string
}

func AddFungibleTokenVaultBatchTransaction(tokens []FungibleTokenInfo) (string, error) {
	return executeTemplate("AddFungibleTokens", AddFungibleTokenVaultBatchTransactionTemplate, tokens)
}

func CreateAccountAndSetupTransaction(tokens []FungibleTokenInfo) (string, error) {
	return executeTemplate("CreateAccount", CreateAccountAndSetupTransactionTemplate, tokens)
}

const CreateAccountAndSetupTransactionTemplate = `
import Crypto
{{ range . }}
import {{ .ContractName }} from {{ .Address }}
{{ end }}

transaction(publicKeys: [Crypto.KeyListEntry]) {
	prepare(signer: AuthAccount) {
		let account = AuthAccount(payer: signer)

		// add all the keys to the account
		for key in publicKeys {
			account.keys.add(publicKey: key.publicKey, hashAlgorithm: key.hashAlgorithm, weight: key.weight)
		}

		{{ range . }}
		// initializing vault for {{ .ContractName }}
		account.save(<-{{ .ContractName }}.createEmptyVault(), to: {{ .VaultStoragePath }})
		account.link<&{{ .ContractName }}.Vault{FungibleToken.Receiver}>(
			{{ .ReceiverPublicPath }},
			target: {{ .VaultStoragePath }}
		)
		account.link<&{{ .ContractName }}.Vault{FungibleToken.Balance}>(
			{{ .BalancePublicPath }},
			target: {{ .VaultStoragePath }}
		)
		{{ end }}
	}
}
`

const AddFungibleTokenVaultBatchTransactionTemplate = `
{{ range . }}
import {{ .ContractName }} from {{ .Address }}
{{ end }}

transaction() {
	prepare(account: AuthAccount) {
		{{ range . }}
		// initializing vault for {{ .ContractName }}
		account.save(<-{{ .ContractName }}.createEmptyVault(), to: {{ .VaultStoragePath }})
		account.link<&{{ .ContractName }}.Vault{FungibleToken.Receiver}>(
			{{ .ReceiverPublicPath }},
			target: {{ .VaultStoragePath }}
		)
		account.link<&{{ .ContractName }}.Vault{FungibleToken.Balance}>(
			{{ .BalancePublicPath }},
			target: {{ .VaultStoragePath }}
		)
		{{ end }}
	}
}
`

func executeTemplate(name string, temp string, tokens []FungibleTokenInfo) (string, error) {
	template, err := template.
		New(name).
		Parse(temp)
	if err != nil {
		return "", nil
	}

	buf := new(bytes.Buffer)
	err = template.Execute(buf, tokens)
	if err != nil {
		return "", nil
	}

	return buf.String(), nil
}
