package tokens

import (
	"github.com/eqlabs/flow-wallet-api/templates"
	"github.com/eqlabs/flow-wallet-api/transactions"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GormStore struct {
	db *gorm.DB
}

func NewGormStore(db *gorm.DB) *GormStore {
	db.AutoMigrate(&AccountToken{}, &FungibleTokenTransfer{})
	return &GormStore{db}
}

func (s *GormStore) AccountTokens(address string, tType *templates.TokenType) (att []AccountToken, err error) {
	q := s.db
	if tType != nil {
		// Filter by type
		q = q.Where(&AccountToken{AccountAddress: address, TokenType: *tType})
	} else {
		// Find all
		q = q.Where(&AccountToken{AccountAddress: address})
	}
	err = q.Order("token_name asc").Find(&att).Error
	return
}

func (s *GormStore) InsertAccountToken(at *AccountToken) error {
	// FirstOrCreate as that will just return the first match instead of throwing
	// a duplicate key error
	return s.db.FirstOrCreate(&AccountToken{}, at).Error
}

func (s *GormStore) InsertFungibleTokenTransfer(t *FungibleTokenTransfer) error {
	return s.db.Create(t).Error
}

// TODO: DRY

func (s *GormStore) FungibleTokenWithdrawals(address, tokenName string) (tt []*FungibleTokenTransfer, err error) {
	tType := transactions.FtTransfer // This needs to be here separately
	err = s.db.
		Preload(clause.Associations).
		Select("*").
		Joins("left join transactions on fungible_token_transfers.transaction_id = transactions.transaction_id").
		Where("transactions.transaction_type = ?", tType).
		Where("transactions.payer_address = ?", address).
		Where("fungible_token_transfers.token_name = ?", tokenName).
		Order("fungible_token_transfers.created_at desc").
		Find(&tt).Error
	return
}

func (s *GormStore) FungibleTokenWithdrawal(address, tokenName, transactionId string) (t *FungibleTokenTransfer, err error) {
	tType := transactions.FtTransfer // This needs to be here separately
	err = s.db.
		Preload(clause.Associations).
		Select("*").
		Joins("left join transactions on fungible_token_transfers.transaction_id = transactions.transaction_id").
		Where("transactions.transaction_type = ?", tType).
		Where("transactions.payer_address = ?", address).
		Where("transactions.transaction_id = ?", transactionId).
		Where("fungible_token_transfers.token_name = ?", tokenName).
		Order("fungible_token_transfers.created_at desc").
		First(&t).Error
	return
}

func (s *GormStore) FungibleTokenDeposits(address, tokenName string) (tt []*FungibleTokenTransfer, err error) {
	tType := transactions.FtTransfer // This needs to be here separately
	err = s.db.
		Preload(clause.Associations).
		Select("*").
		Joins("left join transactions on fungible_token_transfers.transaction_id = transactions.transaction_id").
		Where("transactions.transaction_type = ?", tType).
		Where("fungible_token_transfers.token_name = ?", tokenName).
		Where("fungible_token_transfers.recipient_address = ?", address).
		Order("fungible_token_transfers.created_at desc").
		Find(&tt).Error
	return
}

func (s *GormStore) FungibleTokenDeposit(address, tokenName, transactionId string) (t *FungibleTokenTransfer, err error) {
	tType := transactions.FtTransfer // This needs to be here separately
	err = s.db.
		Preload(clause.Associations).
		Select("*").
		Joins("left join transactions on fungible_token_transfers.transaction_id = transactions.transaction_id").
		Where("fungible_token_transfers.token_name = ?", tokenName).
		Where("transactions.transaction_type = ?", tType).
		Where("transactions.transaction_id = ?", transactionId).
		Where("fungible_token_transfers.recipient_address = ?", address).
		Order("fungible_token_transfers.created_at desc").
		First(&t).Error
	return
}
