package keys

import (
	"gorm.io/gorm"
)

type GormStore struct {
	db *gorm.DB
}

func NewGormStore(db *gorm.DB) *GormStore {
	db.AutoMigrate(&StorableKey{})
	return &GormStore{db}
}

func (s *GormStore) AccountKey(address string) (key StorableKey, err error) {
	err = s.db.Where(&StorableKey{AccountAddress: address}).Order("updated_at asc").First(&key).Error
	s.db.Save(&key) // Update the UpdatedAt field
	return
}
