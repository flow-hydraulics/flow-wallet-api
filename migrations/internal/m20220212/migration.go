package m20220212

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const ID = "20220212"

// State is a type for Job state.
type State string

// Job database model
type Job struct {
	ID                     uuid.UUID      `gorm:"column:id;primary_key;type:uuid;"`
	Type                   string         `gorm:"column:type"`
	State                  State          `gorm:"column:state;default:INIT;index:idx_jobs_state_updated_at"`
	Error                  string         `gorm:"column:error"`
	Errors                 pq.StringArray `gorm:"column:errors;type:text[]"`
	Result                 string         `gorm:"column:result"`
	TransactionID          string         `gorm:"column:transaction_id"`
	ExecCount              int            `gorm:"column:exec_count;default:0"`
	CreatedAt              time.Time      `gorm:"column:created_at"`
	UpdatedAt              time.Time      `gorm:"column:updated_at;index:idx_jobs_state_updated_at"`
	DeletedAt              gorm.DeletedAt `gorm:"column:deleted_at;index"`
	ShouldSendNotification bool           `gorm:"-"` // Whether or not to notify admin (via webhook for example)
	Attributes             datatypes.JSON `gorm:"attributes"`
}

func (Job) TableName() string {
	return "jobs"
}

func Migrate(tx *gorm.DB) error {
	if err := tx.Migrator().CreateIndex(&Job{}, "idx_jobs_state_updated_at"); err != nil {
		return err
	}

	return nil
}

func Rollback(tx *gorm.DB) error {
	if err := tx.Migrator().DropIndex(&Job{}, "idx_jobs_state_updated_at"); err != nil {
		return err
	}

	return nil
}
