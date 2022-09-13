package ops

func (s ServiceImpl) GetMissingFungibleTokenVaults() ([]TokenCount, error) {

	res, err := s.store.ListAccountsWithMissingVault("USDC")
	if err != nil {
		return nil, err
	}

	return []TokenCount{
		{
			TokenName: "USDC",
			Count:     uint(len(*res)),
		},
	}, nil
}

func (s ServiceImpl) InitMissingFungibleTokenVaults() (bool, error) {
	return false, nil
}
