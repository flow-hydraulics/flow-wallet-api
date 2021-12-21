// m20211221_1 handles adding the `Errors` field to Job
package m20211221_1

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const ID = "20211221_1"

// State is a type for Job state.
type State string

// Job database model
type Job struct {
	ID                     uuid.UUID      `gorm:"column:id;primary_key;type:uuid;"`
	Type                   string         `gorm:"column:type"`
	State                  State          `gorm:"column:state;default:INIT"`
	Error                  string         `gorm:"column:error"`
	Errors                 pq.StringArray `gorm:"column:errors;type:text[]"`
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
	if err := tx.Migrator().DropColumn(&Job{}, "errors"); err != nil {
		return err
	}

	return nil
}
