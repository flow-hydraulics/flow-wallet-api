package keys

import (
	"gorm.io/gorm"
)

type GormStore struct {
	db *gorm.DB
}

func NewGormStore(db *gorm.DB) *GormStore {
	db.AutoMigrate(&Storable{})
	return &GormStore{db}
}

func (s *GormStore) AccountKey(address string) (k Storable, err error) {
	err = s.db.Where(&Storable{AccountAddress: address}).Order("updated_at asc").First(&k).Error
	s.db.Save(&k) // Update the UpdatedAt field
	return
}
