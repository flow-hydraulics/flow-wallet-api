package ops

import (
	"context"
	"fmt"
	"sync"

	"github.com/flow-hydraulics/flow-wallet-api/templates"
	"github.com/flow-hydraulics/flow-wallet-api/templates/template_strings"
	"github.com/flow-hydraulics/flow-wallet-api/transactions"
	log "github.com/sirupsen/logrus"
)

type TokenCount struct {
	TokenName string `json:"token"`
	Count     uint   `json:"count"`
}

// GetMissingFungibleTokenVaults returns number of accounts that are missing a configured fungible token vault.
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

// InitMissingFungibleTokenVaults starts job to init missing fungible token vaults.
func (s *ServiceImpl) InitMissingFungibleTokenVaults() (string, error) {
	if s.initFungibleJobRunning {
		return "Job is already running", nil
	}

	s.initFungibleJobRunning = true

	log.Info("Starting new fungible token vault init job")

	tokens, err := s.temps.ListTokensFull(templates.FT)
	if err != nil {
		return "", err
	}

	// mapping of user account address to list of fungible tokens that needs to be initialized
	accountsMap := make(map[string][]string)
	// token name -> FungibleTokenInfo for templating
	tokenInfoMap := make(map[string]template_strings.FungibleTokenInfo)
	for _, t := range tokens {
		if t.Name != "FlowToken" {
			tokenInfoMap[t.Name] = templates.NewFungibleTokenInfo(t)

			accounts, err := s.store.ListAccountsWithMissingVault(t.Name)
			if err != nil {
				return "", err
			}
			for _, a := range *accounts {
				accountsMap[a.Address] = append(accountsMap[a.Address], t.Name)
			}
		}
	}

	log.Infof("Number of accounts for vault init job: %d", len(accountsMap))

	var txWg sync.WaitGroup

	for address, tokenList := range accountsMap {

		txWg.Add(1)

		// create a new job for each account to init token vaults
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

					txScript, err := templates.InitFungibleTokenVaultsCode(s.cfg.ChainID, tList)
					if err != nil {
						return err
					}

					// blocks until transaction is sealed
					_, tx, err := s.txs.Create(context.Background(), true, address, txScript, nil, transactions.FtSetup)
					if err != nil {
						return err
					}

					for _, t := range tokenList {
						err := s.tokens.AddAccountToken(t, address)
						if err != nil {
							log.Errorf("Error adding AccountToken to store: %s", err)
						}
					}

					log.
						WithFields(log.Fields{
							"account": address,
							"txId":    tx.TransactionId,
							"tokens":  tokenList,
						}).
						Info("Fungible token vaults initialized.")

					return nil
				},
			},
		)
	}

	// unlock after all transactions are sealed
	go func() {
		txWg.Wait()
		s.initFungibleJobRunning = false
		log.Info("Fungible token vault init job finished")
	}()

	return fmt.Sprintf("Job started! Accounts: %d, Workers: %d", len(accountsMap), s.wp.NumWorkers()), nil
}
