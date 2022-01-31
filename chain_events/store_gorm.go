package chain_events

import (
	"sync"

	"github.com/flow-hydraulics/flow-wallet-api/datastore/lib"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GormStore struct {
	statusMutex sync.Mutex
	db          *gorm.DB
}

func NewGormStore(db *gorm.DB) Store {
	return &GormStore{db: db}
}

// LockedStatus runs a transaction on the database manipulating 'status' of type ListenerStatus.
func (s *GormStore) LockedStatus(fn func(status *ListenerStatus) error) error {
	s.statusMutex.Lock()
	defer s.statusMutex.Unlock()

	return lib.GormTransaction(s.db, func(tx *gorm.DB) error {
		status := ListenerStatus{}

		if err := tx.
			// NOWAIT so this call will fail rather than use a stale value
			Clauses(clause.Locking{Strength: "UPDATE", Options: "NOWAIT"}).
			FirstOrCreate(&status).Error; err != nil {
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
