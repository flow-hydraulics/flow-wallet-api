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

func (s *GormStore) ProposalKey() (i int, err error) {
	p := ProposalKey{}
	err = s.db.Model(&ProposalKey{}).Order("updated_at asc").First(&p).Error
	s.db.Save(&p) // Update the UpdatedAt field
	i = p.KeyIndex
	return
}

func (s *GormStore) InsertProposalKey(p ProposalKey) error {
	return s.db.Create(&p).Error
}

func (s *GormStore) DeleteAllProposalKeys() error {
	return s.db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&ProposalKey{}).Error
}
