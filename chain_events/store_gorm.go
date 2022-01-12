package chain_events

import (
	"gorm.io/gorm"
)

type GormStore struct {
	db *gorm.DB
}

func NewGormStore(db *gorm.DB) Store {
	return &GormStore{db}
}

// LockedStatus runs a transaction on the database manipulating 'status' of type ListenerStatus.
func (s *GormStore) LockedStatus(fn func(status *ListenerStatus) error) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		status := ListenerStatus{}

		if err := tx.FirstOrCreate(&status).Error; err != nil {
			return err // rollback
		}

		if err := fn(&status); err != nil {
			return err // rollback
		}

		if err := tx.Save(&status).Error; err != nil {
			return err // rollback
		}

		return nil // commit
	})
}
