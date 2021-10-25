package m20211005

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const ID = "20211005"

type Job struct {
	ID            uuid.UUID      `gorm:"column:id;primary_key;type:uuid;"`
	Status        int            `gorm:"column:status"`
	State         string         `gorm:"column:state"`
	Type          string         `gorm:"column:type"`
	Error         string         `gorm:"column:error"`
	Result        string         `gorm:"column:result"`
	ExecCount     int            `gorm:"column:exec_count;default:0"`
	TransactionID string         `gorm:"column:transaction_id"`
	CreatedAt     time.Time      `gorm:"column:created_at"`
	UpdatedAt     time.Time      `gorm:"column:updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func Migrate(tx *gorm.DB) error {
	if err := tx.AutoMigrate(&Job{}); err != nil {
		return err
	}

	/*
		if err := tx.Model(&Job{}).Where("status = ?", 0).Update("state", "INIT"); err != nil {
			return err
		}

		if err := tx.Model(&Job{}).Where("status = ?", 1).Update("state", "INIT"); err != nil {
			return err
		}

		if err := tx.Model(&Job{}).Where("status = ?", 2).Update("state", "ACCEPTED"); err != nil {
			return err
		}

		if err := tx.Model(&Job{}).Where("status = ?", 3).Update("state", "NO_AVAILABLE_WORKERS"); err != nil {
			return err
		}

		if err := tx.Model(&Job{}).Where("status = ?", 4).Update("state", "NO_AVAILABLE_WORKERS"); err != nil {
			return err
		}

		if err := tx.Model(&Job{}).Where("status = ?", 5).Update("state", "FAILED"); err != nil {
			return err
		}

		if err := tx.Model(&Job{}).Where("status = ?", 6).Update("state", "COMPLETE"); err != nil {
			return err
		}

		if err := tx.Model(&Job{}).Where("status > ?", 6).Update("state", "FAILED"); err != nil {
			return err
		}

		if err := tx.DropColumn(&Job{}, "status"); err != nil {
			return err
		}
	*/

	err := tx.Model(&Job{}).
		Where("status = ?", 0).Update("state", "INIT").
		Where("status = ?", 1).Update("state", "INIT").
		Where("status = ?", 2).Update("state", "ACCEPTED").
		Where("status = ?", 3).Update("state", "NO_AVAILABLE_WORKERS").
		Where("status = ?", 4).Update("state", "NO_AVAILABLE_WORKERS").
		Where("status = ?", 5).Update("state", "FAILED").
		Where("status = ?", 6).Update("state", "COMPLETE").
		Where("status > ?", 6).Update("state", "FAILED").Error

	if err != nil {
		return err
	}

	if err = tx.Migrator().DropColumn(&Job{}, "status"); err != nil {
		return err
	}

	return nil
}

func Rollback(tx *gorm.DB) error {
	if err := tx.AutoMigrate(&Job{}); err != nil {
		return err
	}

	err := tx.Model(&Job{}).
		Where("state = ?", "INIT").Update("status", 0).
		Where("state = ?", "INIT").Update("status", 1).
		Where("state = ?", "ACCEPTED").Update("status", 2).
		Where("state = ?", "NO_AVAILABLE_WORKERS").Update("status", 3).
		Where("state = ?", "FAILED").Update("status", 5).
		Where("state = ?", "COMPLETE").Update("status", 6).Error

	if err != nil {
		return err
	}

	if err := tx.Migrator().DropColumn(&Job{}, "status"); err != nil {
		return err
	}

	if err := tx.Migrator().DropColumn(&Job{}, "type"); err != nil {
		return err
	}

	if err := tx.Migrator().DropColumn(&Job{}, "exec_count"); err != nil {
		return err
	}

	return nil
}
