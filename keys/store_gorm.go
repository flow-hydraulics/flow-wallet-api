package keys

import (
	"gorm.io/gorm"
)

type GormStore struct {
	db *gorm.DB
}

func NewGormStore(db *gorm.DB) *GormStore {
	return &GormStore{db}
}

func (s *GormStore) AccountKey(address string) (k Storable, err error) {
	err = s.db.Where(&Storable{AccountAddress: address}).Order("updated_at asc").First(&k).Error
	s.db.Save(&k) // Update the UpdatedAt field
	return
}

func (s *GormStore) ProposalKey() (int, error) {
	p := ProposalKey{}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&ProposalKey{}).Order("updated_at asc").First(&p).Error; err != nil {
			return err
		}

		if err := tx.Save(&p).Error; err != nil {
			return err
		}

		return nil
	})

	return p.KeyIndex, err
}

func (s *GormStore) InsertProposalKey(p ProposalKey) error {
	return s.db.Create(&p).Error
}

func (s *GormStore) DeleteAllProposalKeys() error {
	return s.db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&ProposalKey{}).Error
}
