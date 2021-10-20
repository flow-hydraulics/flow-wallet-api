package m20210922

import (
	"time"

	"github.com/flow-hydraulics/flow-wallet-api/templates"
	"github.com/flow-hydraulics/flow-wallet-api/transactions"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

//
// This is the first migration that initializes the whole DB.  All types are
// snapshot here so that the structure and schema state for given point in time
// is preserved and can be rolled back to from later migrations, in case
// there's a need.
//

const ID = "20210922"

type ListenerStatus struct {
	gorm.Model
	LatestHeight uint64
}

func (ListenerStatus) TableName() string {
	return "chain_events_status"
}

type Job struct {
	ID            uuid.UUID      `gorm:"column:id;primary_key;type:uuid;"`
	Status        int            `gorm:"column:status"`
	Error         string         `gorm:"column:error"`
	Result        string         `gorm:"column:result"`
	TransactionID string         `gorm:"column:transaction_id"`
	CreatedAt     time.Time      `gorm:"column:created_at"`
	UpdatedAt     time.Time      `gorm:"column:updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

type Storable struct {
	ID             int    `gorm:"primaryKey"`
	AccountAddress string `gorm:"index"`
	Index          int    `gorm:"index"`
	Type           string
	Value          []byte
	SignAlgo       string
	HashAlgo       string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"`
}

func (Storable) TableName() string {
	return "storable_keys"
}

type ProposalKey struct {
	ID        int `gorm:"primaryKey"`
	KeyIndex  int `gorm:"unique"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (ProposalKey) TableName() string {
	return "proposal_keys"
}

type Account struct {
	Address   string     `gorm:"primaryKey"`
	Keys      []Storable `gorm:"foreignKey:AccountAddress;references:Address;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Type      string     `gorm:"default:custodial"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type Token struct {
	ID            uint64
	Name          string `gorm:"uniqueIndex;not null"`
	NameLowerCase string
	Address       string `gorm:"not null"`
	Setup         string
	Transfer      string
	Balance       string
	Type          int
}

type AccountToken struct {
	ID             uint64              `json:"-" gorm:"column:id;primaryKey"`
	AccountAddress string              `json:"-" gorm:"column:account_address;uniqueIndex:addressname;index;not null"`
	Account        Account             `json:"-" gorm:"foreignKey:AccountAddress;references:Address;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	TokenName      string              `json:"name" gorm:"column:token_name;uniqueIndex:addressname;index;not null"`
	TokenAddress   string              `json:"address" gorm:"column:token_address;uniqueIndex:addressname;index;not null"`
	TokenType      templates.TokenType `json:"-" gorm:"column:token_type"`
	CreatedAt      time.Time           `json:"-" gorm:"column:created_at"`
	UpdatedAt      time.Time           `json:"-" gorm:"column:updated_at"`
	DeletedAt      gorm.DeletedAt      `json:"-" gorm:"column:deleted_at;index"`
}

func (AccountToken) TableName() string {
	return "account_tokens"
}

type TokenTransfer struct {
	ID               uint64                   `gorm:"column:id;primaryKey"`
	TransactionId    string                   `gorm:"column:transaction_id"`
	Transaction      transactions.Transaction `gorm:"foreignKey:TransactionId;references:TransactionId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	RecipientAddress string                   `gorm:"column:recipient_address;index"`
	SenderAddress    string                   `gorm:"column:sender_address;index"`
	FtAmount         string                   `gorm:"column:ft_amount"`
	NftID            uint64                   `gorm:"column:nft_id"`
	TokenName        string                   `gorm:"column:token_name"`
	CreatedAt        time.Time                `gorm:"column:created_at"`
	UpdatedAt        time.Time                `gorm:"column:updated_at"`
	DeletedAt        gorm.DeletedAt           `gorm:"column:deleted_at;index"`
}

func (TokenTransfer) TableName() string {
	return "token_transfers"
}

type Transaction struct {
	TransactionId   string         `gorm:"column:transaction_id;primaryKey"`
	TransactionType int            `gorm:"column:transaction_type;index"`
	ProposerAddress string         `gorm:"column:proposer_address;index"`
	CreatedAt       time.Time      `gorm:"column:created_at"`
	UpdatedAt       time.Time      `gorm:"column:updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (Transaction) TableName() string {
	return "transactions"
}

func Migrate(tx *gorm.DB) error {
	tx.AutoMigrate(&ListenerStatus{})
	tx.AutoMigrate(&Job{})

	// Storables has had migration in the past. Just keep it as-is.
	tx.Migrator().RenameTable("storables", "storable_keys") // Ignore error
	tx.AutoMigrate(&ProposalKey{}, &Storable{})

	tx.Migrator().RenameTable("fungible_token_transfers", "token_transfers")
	tx.AutoMigrate(&AccountToken{}, &TokenTransfer{})

	tx.Migrator().RenameColumn(&Transaction{}, "payer_address", "proposer_address")
	tx.AutoMigrate(&Transaction{})

	tx.AutoMigrate(&Account{})
	tx.AutoMigrate(&Token{})

	// Migrating from transaction.payer_address to transaction.proposer_address
	// This change meant that transactions payer or proposer no longer equals
	// the actual sender of a token transfer. From now on token transfers have
	// their own sender_address column and thus there may be old token transfers whose
	// sender_address is NULL.
	// This migration updates sender_address columns which are NULL to equal
	// the transactions proposer_address. This assumption is ok as sender_address
	// column should be NULL only when this migration is run the first time.
	tx.Table("token_transfers as tt").
		Where(map[string]interface{}{"sender_address": nil}).
		Update("sender_address", tx.Table("transactions as tx").
			Select("proposer_address").
			Where("tx.transaction_id = tt.transaction_id"))

	return nil
}

func Rollback(tx *gorm.DB) error {
	if err := tx.Migrator().DropTable(&Token{}); err != nil {
		return err
	}

	if err := tx.Migrator().DropTable(&Account{}); err != nil {
		return err
	}

	if err := tx.Migrator().DropTable(&Transaction{}); err != nil {
		return err
	}

	if err := tx.Migrator().DropTable(&AccountToken{}, &TokenTransfer{}); err != nil {
		return err
	}

	if err := tx.Migrator().DropTable(&ProposalKey{}, &Storable{}); err != nil {
		return err
	}

	if err := tx.Migrator().DropTable(&Job{}); err != nil {
		return err
	}

	if err := tx.Migrator().DropTable(&ListenerStatus{}); err != nil {
		return err
	}

	return nil
}
