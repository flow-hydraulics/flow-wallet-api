package templates

import "github.com/flow-hydraulics/flow-wallet-api/templates/template_strings"

func NewFungibleTokenInfo(t Token) template_strings.FungibleTokenInfo {
	return template_strings.FungibleTokenInfo{
		ContractName:       t.Name,
		Address:            t.Address,
		VaultStoragePath:   t.VaultStoragePath,
		ReceiverPublicPath: t.ReceiverPublicPath,
		BalancePublicPath:  t.BalancePublicPath,
	}
}
