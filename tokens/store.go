package tokens

// Store manages data regarding tokens.
type Store interface {
	InsertFungibleTokenTransfer(*FungibleTokenTransfer) error
	FungibleTokenWithdrawals(address, tokenName string) ([]*FungibleTokenTransfer, error)
	FungibleTokenWithdrawal(address, tokenName, transactionId string) (*FungibleTokenTransfer, error)
	FungibleTokenDeposits(address, tokenName string) ([]*FungibleTokenTransfer, error)
	FungibleTokenDeposit(address, tokenName, transactionId string) (*FungibleTokenTransfer, error)
}
