package ops

import (
	"context"

	"github.com/flow-hydraulics/flow-wallet-api/templates"
	"github.com/flow-hydraulics/flow-wallet-api/templates/template_strings"
	"github.com/flow-hydraulics/flow-wallet-api/transactions"
	log "github.com/sirupsen/logrus"
)

func (s ServiceImpl) GetMissingFungibleTokenVaults() ([]TokenCount, error) {

	tokens, err := s.temps.ListTokensFull(templates.FT)
	if err != nil {
		return nil, err
	}

	var result []TokenCount
	for _, t := range tokens {
		if t.Name != "FlowToken" {
			accounts, err := s.store.ListAccountsWithMissingVault(t.Name)
			if err != nil {
				return nil, err
			}
			result = append(result, TokenCount{
				TokenName: t.Name,
				Count:     uint(len(*accounts)),
			})
		}
	}

	return result, nil
}

func (s ServiceImpl) InitMissingFungibleTokenVaults() (bool, error) {
	ctx := context.Background()
	tokens, err := s.temps.ListTokensFull(templates.FT)
	if err != nil {
		return false, err
	}

	accountsMap := make(map[string][]string)
	tokenInfoMap := make(map[string]template_strings.FungibleTokenInfo)
	for _, t := range tokens {
		if t.Name != "FlowToken" {
			tokenInfoMap[t.Name] = template_strings.FungibleTokenInfo{
				ContractName:       t.Name,
				Address:            t.Address,
				VaultStoragePath:   t.VaultStoragePath,
				ReceiverPublicPath: t.ReceiverPublicPath,
				BalancePublicPath:  t.BalancePublicPath,
			}

			accounts, err := s.store.ListAccountsWithMissingVault(t.Name)
			if err != nil {
				return false, err
			}
			for _, a := range *accounts {
				accountsMap[a.Address] = append(accountsMap[a.Address], t.Name)
			}
		}
	}

	for address, tokenList := range accountsMap {

		tList := []template_strings.FungibleTokenInfo{}
		for _, t := range tokenList {
			tList = append(tList, tokenInfoMap[t])
		}

		txScript, err := templates.AddFungibleTokenVaultBatchCode(s.cfg.ChainID, tList)
		if err != nil {
			return false, err
		}

		job, _, err := s.txs.Create(ctx, false, address, txScript, nil, transactions.FtSetup)
		if err != nil {
			return false, err
		}

		log.Debug(job)

	}

	return false, nil
}
