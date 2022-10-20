package ops

import (
	"github.com/flow-hydraulics/flow-wallet-api/accounts"
	"github.com/flow-hydraulics/flow-wallet-api/templates"
	"github.com/flow-hydraulics/flow-wallet-api/tokens"
	"gorm.io/gorm"
)

type GormStore struct {
	db *gorm.DB
}

func NewGormStore(db *gorm.DB) Store {
	return &GormStore{db}
}

// ListAccountsWithMissingVault lists accounts that are not initialized for the given token name.
func (s *GormStore) ListAccountsWithMissingVault(tokenName string) (res *[]accounts.Account, err error) {

	// https://www.db-fiddle.com/f/hHoZ5P4FDDj3sVkRRiRMc2/0
	err = s.db.
		Where(
			"not exists (?)",
			s.db.
				Model(&tokens.AccountToken{}).
				Where(&tokens.AccountToken{
					TokenName: tokenName,
					TokenType: templates.FT,
				}).
				Where("account_address=address")).
		Find(&res).Error

	return
}
