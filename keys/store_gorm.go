package keys

import (
	"gorm.io/gorm"
)

type GormStore struct {
	db *gorm.DB
}

func NewGormStore(db *gorm.DB) *GormStore {
	db.Migrator().RenameTable("storables", "storable_keys") // Ignore error
	db.AutoMigrate(&Storable{})
	db.AutoMigrate(&Proposer{})
	return &GormStore{db}
}

func (s *GormStore) AccountKey(address string) (k Storable, err error) {
	err = s.db.Where(&Storable{AccountAddress: address}).Order("updated_at asc").First(&k).Error
	s.db.Save(&k) // Update the UpdatedAt field
	return
}

func (s *GormStore) Proposer() (i int, err error) {
	p := Proposer{}
	err = s.db.Model(&Proposer{}).Order("updated_at asc").First(&p).Error
	s.db.Save(&p) // Update the UpdatedAt field
	i = p.KeyIndex
	return
}

func (s *GormStore) InsertProposer(p Proposer) error {
	return s.db.Create(&p).Error
}

func (s *GormStore) DeleteAllProposers() error {
	return s.db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&Proposer{}).Error
}
