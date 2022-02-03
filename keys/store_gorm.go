package keys

import (
	"sync"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/flow-hydraulics/flow-wallet-api/datastore/lib"
)

type GormStore struct {
	accountKeyMutex  sync.Mutex
	proposalKeyMutex sync.Mutex
	db               *gorm.DB
}

func NewGormStore(db *gorm.DB) Store {
	return &GormStore{db: db}
}

func (s *GormStore) AccountKey(address string) (Storable, error) {
	s.accountKeyMutex.Lock()
	defer s.accountKeyMutex.Unlock()

	k := Storable{}

	err := lib.GormTransaction(s.db, func(tx *gorm.DB) error {
		if err := tx.
			// NOWAIT so this call will fail rather than use a stale value
			Clauses(clause.Locking{Strength: "UPDATE", Options: "NOWAIT"}).
			Where(&Storable{AccountAddress: address}).
			Order("updated_at asc").
			Limit(1).Find(&k).Error; err != nil {
			return err
		}

		if err := tx.Model(&k).Update("updated_at", time.Now()).Error; err != nil {
			return err
		}

		return nil
	})

	return k, err
}

func (s *GormStore) ProposalKeyIndex(limitKeyCount int) (int, error) {
	s.proposalKeyMutex.Lock()
	defer s.proposalKeyMutex.Unlock()

	p := ProposalKey{}

	err := lib.GormTransaction(s.db, func(tx *gorm.DB) error {
		if err := tx.Table("(?) as p", tx.Model(p).Order("id asc").Limit(limitKeyCount)).
			// NOWAIT so this call will fail rather than use a stale value
			Clauses(clause.Locking{Strength: "UPDATE", Options: "NOWAIT"}).
			Order("updated_at asc").
			Limit(1).Find(&p).Error; err != nil {
			return err
		}

		if err := tx.Model(&p).Update("updated_at", time.Now()).Error; err != nil {
			return err
		}

		return nil
	})

	return p.KeyIndex, err
}

func (s *GormStore) ProposalKeyCount() (int64, error) {
	var count int64
	return count, s.db.Table(ProposalKey{}.TableName()).Count(&count).Error
}

func (s *GormStore) InsertProposalKey(p ProposalKey) error {
	return s.db.Create(&p).Error
}

func (s *GormStore) DeleteAllProposalKeys() error {
	return s.db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&ProposalKey{}).Error
}
