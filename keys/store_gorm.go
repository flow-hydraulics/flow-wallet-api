package keys

import (
	"time"

	"gorm.io/gorm"
)

type GormStore struct {
	db *gorm.DB
}

func NewGormStore(db *gorm.DB) *GormStore {
	return &GormStore{db}
}

func (s *GormStore) AccountKey(address string) (Storable, error) {
	k := Storable{}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.
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

func (s *GormStore) ProposalKeyIndex() (int, error) {
	p := ProposalKey{}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.
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
