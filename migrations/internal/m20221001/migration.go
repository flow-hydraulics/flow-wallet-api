package m20221001

import (
	"gorm.io/gorm"
)

const ID = "20221001"

type TokenType int

const (
	NotSpecified TokenType = iota
	FT
	NFT
)

type Token struct {
	ID                 uint64    `json:"id,omitempty"`
	Name               string    `json:"name" gorm:"uniqueIndex;not null"` // Declaration name
	NameLowerCase      string    `json:"nameLowerCase,omitempty"`          // (deprecated) For generic fungible token transaction templates
	ReceiverPublicPath string    `json:"receiverPublicPath,omitempty"`
	BalancePublicPath  string    `json:"balancePublicPath,omitempty"`
	VaultStoragePath   string    `json:"vaultStoragePath,omitempty"`
	Address            string    `json:"address" gorm:"not null"`
	Setup              string    `json:"setup,omitempty"`    // Setup cadence code
	Transfer           string    `json:"transfer,omitempty"` // Transfer cadence code
	Balance            string    `json:"balance,omitempty"`  // Balance cadence code
	Type               TokenType `json:"type"`
}

func Migrate(tx *gorm.DB) error {
	if err := tx.Migrator().AddColumn(&Token{}, "receiver_public_path"); err != nil {
		return err
	}

	if err := tx.Migrator().AddColumn(&Token{}, "balance_public_path"); err != nil {
		return err
	}

	if err := tx.Migrator().AddColumn(&Token{}, "vault_storage_path"); err != nil {
		return err
	}

	return nil
}

func Rollback(tx *gorm.DB) error {
	if err := tx.Migrator().DropColumn(&Token{}, "receiver_public_path"); err != nil {
		return err
	}

	if err := tx.Migrator().DropColumn(&Token{}, "balance_public_path"); err != nil {
		return err
	}

	if err := tx.Migrator().DropColumn(&Token{}, "vault_storage_path"); err != nil {
		return err
	}

	return nil
}
