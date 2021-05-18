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

func (s *GormStore) AccountKey(address string) (key Storable, err error) {
	err = s.db.Where(&Storable{AccountAddress: address}).Order("updated_at asc").First(&key).Error
	s.db.Save(&key) // Update the UpdatedAt field
	return
}
