package templates

import (
	"gorm.io/gorm"
)

type GormStore struct {
	db *gorm.DB
}

func NewGormStore(db *gorm.DB) *GormStore {
	db.AutoMigrate(&Token{})
	return &GormStore{db}
}

func (s *GormStore) Insert(q *Token) error {
	return s.db.Omit("ID").Create(q).Error
}

func (s *GormStore) List(tType *TokenType) (*[]Token, error) {
	var tt = &[]Token{}
	var err error

	if tType != nil {
		// Filter by type
		err = s.db.Where(&Token{Type: *tType}).Find(tt).Error
	} else {
		// Find all
		err = s.db.Find(tt).Error
	}

	if err != nil {
		return nil, err
	}

	return tt, nil
}

func (s *GormStore) Get(id uint64) (*Token, error) {
	var token Token
	err := s.db.First(&token, id).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (s *GormStore) Remove(id uint64) error {
	return s.db.Delete(&Token{}, id).Error
}
