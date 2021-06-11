package tokens

import (
	"github.com/eqlabs/flow-wallet-service/transactions"
	"gorm.io/gorm"
)

type GormStore struct {
	db *gorm.DB
}

func NewGormStore(db *gorm.DB) *GormStore {
	db.AutoMigrate(&FungibleTokenTransfer{})
	return &GormStore{db}
}

func (s *GormStore) InsertFungibleTokenTransfer(t *FungibleTokenTransfer) error {
	return s.db.Create(t).Error
}

func (s *GormStore) FungibleTokenWithdrawals(address, tokenName string) (tt []FungibleTokenTransfer, err error) {
	tType := transactions.FtWithdrawal // This needs to be here separately
	err = s.db.
		Select("*").
		Joins("left join transactions on fungible_token_transfers.transaction_id = transactions.transaction_id").
		Where("fungible_token_transfers.token_name = ?", tokenName).
		Where("transactions.transaction_type = ?", tType).
		Where("transactions.payer_address = ?", address).
		Order("fungible_token_transfers.created_at desc").
		Find(&tt).Error
	return
}

func (s *GormStore) FungibleTokenWithdrawal(address, tokenName, transactionId string) (t FungibleTokenTransfer, err error) {
	tType := transactions.FtWithdrawal // This needs to be here separately
	err = s.db.Select("*").
		Joins("left join transactions on fungible_token_transfers.transaction_id = transactions.transaction_id").
		Where("fungible_token_transfers.token_name = ?", tokenName).
		Where("transactions.transaction_id = ?", transactionId).
		Where("transactions.transaction_type = ?", tType).
		Where("transactions.payer_address = ?", address).
		Order("fungible_token_transfers.created_at desc").
		First(&t).Error
	return
}
