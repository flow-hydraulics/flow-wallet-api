// m20211220 handles Job.Attributes migration
package m20211220

import (
	"time"

	"github.com/flow-hydraulics/flow-wallet-api/jobs"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const ID = "20211220"

// Job database model
type Job struct {
	ID                     uuid.UUID      `gorm:"column:id;primary_key;type:uuid;"`
	Type                   string         `gorm:"column:type"`
	State                  jobs.State     `gorm:"column:state;default:INIT"`
	Error                  string         `gorm:"column:error"`
	Result                 string         `gorm:"column:result"`
	TransactionID          string         `gorm:"column:transaction_id"`
	ExecCount              int            `gorm:"column:exec_count;default:0"`
	CreatedAt              time.Time      `gorm:"column:created_at"`
	UpdatedAt              time.Time      `gorm:"column:updated_at"`
	DeletedAt              gorm.DeletedAt `gorm:"column:deleted_at;index"`
	ShouldSendNotification bool           `gorm:"-"` // Whether or not to notify admin (via webhook for example)
	Attributes             datatypes.JSON `gorm:"attributes"`
}

func (Job) TableName() string {
	return "jobs"
}

func Migrate(tx *gorm.DB) error {
	if err := tx.AutoMigrate(&Job{}); err != nil {
		return err
	}

	return nil
}

func Rollback(tx *gorm.DB) error {
	if err := tx.Migrator().DropColumn(&Job{}, "attributes"); err != nil {
		return err
	}

	return nil
}
