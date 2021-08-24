package tokens

import (
	"fmt"

	"github.com/flow-hydraulics/flow-wallet-api/templates"
	"github.com/flow-hydraulics/flow-wallet-api/transactions"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GormStore struct {
	db *gorm.DB
}

func NewGormStore(db *gorm.DB) *GormStore {
	// Ignoring migration errors

	db.Migrator().RenameTable("fungible_token_transfers", "token_transfers")

	db.AutoMigrate(&AccountToken{}, &TokenTransfer{})

	// Migrating from transaction.payer_address to transaction.proposer_address
	// This change meant that transactions payer or proposer no longer equals
	// the actual sender of a token transfer. From now on token transfers have
	// their own sender_address column and thus there may be old token transfers whose
	// sender_address is NULL.
	// This migration updates sender_address columns which are NULL to equal
	// the transactions proposer_address. This assumption is ok as sender_address
	// column should be NULL only when this migration is run the first time.
	db.Table("token_transfers as tt").
		Where(map[string]interface{}{"sender_address": nil}).
		Update("sender_address", db.Table("transactions as tx").
			Select("proposer_address").
			Where("tx.transaction_id = tt.transaction_id"))

	return &GormStore{db}
}

func (s *GormStore) AccountTokens(address string, tokenType templates.TokenType) (att []AccountToken, err error) {
	q := s.db
	if tokenType != templates.NotSpecified {
		// Filter by type
		q = q.Where(&AccountToken{AccountAddress: address, TokenType: tokenType})
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

func (s *GormStore) InsertTokenTransfer(t *TokenTransfer) error {
	return s.db.Create(t).Error
}

func tokenToTransferType(token *templates.Token) (*transactions.Type, error) {
	var txType transactions.Type
	switch token.Type {
	case templates.FT:
		txType = transactions.FtTransfer
	case templates.NFT:
		txType = transactions.NftTransfer
	default:
		return nil, fmt.Errorf("unknown token type")
	}
	return &txType, nil
}

// TODO: DRY

func (s *GormStore) TokenWithdrawals(address string, token *templates.Token) (tt []*TokenTransfer, err error) {
	txType, err := tokenToTransferType(token)
	if err != nil {
		return nil, err
	}

	err = s.db.
		Preload(clause.Associations).
		Select("*").
		Joins("left join transactions on token_transfers.transaction_id = transactions.transaction_id").
		Where("token_transfers.sender_address = ?", address).
		Where("transactions.transaction_type = ?", txType).
		Where("token_transfers.token_name = ?", token.Name).
		Order("token_transfers.created_at desc").
		Find(&tt).Error
	return
}

func (s *GormStore) TokenWithdrawal(address, transactionId string, token *templates.Token) (t *TokenTransfer, err error) {
	txType, err := tokenToTransferType(token)
	if err != nil {
		return nil, err
	}

	err = s.db.
		Preload(clause.Associations).
		Select("*").
		Joins("left join transactions on token_transfers.transaction_id = transactions.transaction_id").
		Where("token_transfers.sender_address = ?", address).
		Where("transactions.transaction_type = ?", txType).
		Where("transactions.transaction_id = ?", transactionId).
		Where("token_transfers.token_name = ?", token.Name).
		Order("token_transfers.created_at desc").
		First(&t).Error
	return
}

func (s *GormStore) TokenDeposits(address string, token *templates.Token) (tt []*TokenTransfer, err error) {
	txType, err := tokenToTransferType(token)
	if err != nil {
		return nil, err
	}

	err = s.db.
		Preload(clause.Associations).
		Select("*").
		Joins("left join transactions on token_transfers.transaction_id = transactions.transaction_id").
		Where("token_transfers.recipient_address = ?", address).
		Where("transactions.transaction_type = ?", txType).
		Where("token_transfers.token_name = ?", token.Name).
		Order("token_transfers.created_at desc").
		Find(&tt).Error
	return
}

func (s *GormStore) TokenDeposit(address, transactionId string, token *templates.Token) (t *TokenTransfer, err error) {
	txType, err := tokenToTransferType(token)
	if err != nil {
		return nil, err
	}

	err = s.db.
		Preload(clause.Associations).
		Select("*").
		Joins("left join transactions on token_transfers.transaction_id = transactions.transaction_id").
		Where("token_transfers.recipient_address = ?", address).
		Where("transactions.transaction_type = ?", txType).
		Where("transactions.transaction_id = ?", transactionId).
		Where("token_transfers.token_name = ?", token.Name).
		Order("token_transfers.created_at desc").
		First(&t).Error
	return
}
