package ops

import (
	"context"
	"sync"
	"time"

	"github.com/flow-hydraulics/flow-wallet-api/templates"
	"github.com/flow-hydraulics/flow-wallet-api/templates/template_strings"
	"github.com/flow-hydraulics/flow-wallet-api/transactions"
	log "github.com/sirupsen/logrus"
)

func (s *ServiceImpl) GetMissingFungibleTokenVaults() ([]TokenCount, error) {

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

func (s *ServiceImpl) InitMissingFungibleTokenVaults() (bool, error) {
	if s.initFungibleJobRunning {
		log.Debug("Fungible token vault init job is already running")
		return true, nil
	}

	s.initFungibleJobRunning = true

	log.Debug("Starting new fungible token vault init job")

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

	log.Debugf("Number of accounts for vault init job: %d", len(accountsMap))

	var txWg sync.WaitGroup

	for address, tokenList := range accountsMap {

		txWg.Add(1)

		s.wp.AddFungibleInitJob(
			OpsInitFungibleVaultsJob{
				Address:   address,
				TokenList: tokenList,
				Func: func(address string, tokenList []string) error {

					defer txWg.Done()

					tList := []template_strings.FungibleTokenInfo{}
					for _, t := range tokenList {
						tList = append(tList, tokenInfoMap[t])
					}

					txScript, err := templates.AddFungibleTokenVaultBatchCode(s.cfg.ChainID, tList)
					if err != nil {
						return err
					}

					_, tx, err := s.txs.Create(context.Background(), true, address, txScript, nil, transactions.FtSetup)
					if err != nil {
						return err
					}

					for _, t := range tokenList {
						err := s.tokens.AddAccountToken(t, address)
						if err != nil {
							log.Debugf("Error adding AccountToken to store: %s", err)
						}
					}

					log.Debugf("ops transaction sent: %s", tx.TransactionId)

					return nil
				},
			},
		)

		time.Sleep(s.cfg.OpsBurstInterval)
	}

	go func() {
		txWg.Wait()
		s.initFungibleJobRunning = false
		log.Debug("Fungible token vault init job finished")
	}()

	return false, nil
}
