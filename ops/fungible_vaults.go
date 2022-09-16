package ops

import "github.com/flow-hydraulics/flow-wallet-api/templates"

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
	return false, nil
}
