package chain_events

import (
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GormStore struct {
	db *gorm.DB
}

func NewGormStore(db *gorm.DB) *GormStore {
	return &GormStore{db}
}

// LockedStatus runs a transaction on the database manipulating 'status' of type ListenerStatus.
// It takes a function 'fn' as argument. In the context of 'fn' 'status' is locked.
// Uses NOWAIT modifier on the lock so simultaneous requests can be ignored.
func (s *GormStore) LockedStatus(fn func(status *ListenerStatus) error) error {
	txErr := s.db.Transaction(func(tx *gorm.DB) error {
		status := ListenerStatus{}

		if err := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "NOWAIT"}).FirstOrCreate(&status).Error; err != nil {
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

	// Need to handle implementation specific error message
	if txErr != nil && strings.Contains(txErr.Error(), "could not obtain lock on row") {
		return &LockError{Err: txErr}
	}

	return txErr
}
