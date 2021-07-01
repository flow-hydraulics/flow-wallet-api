package templates

import (
	"database/sql"

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

func (s *GormStore) List(tType *TokenType) (*[]BasicToken, error) {
	var tt = &[]BasicToken{}
	var err error

	q := s.db.Model(&Token{})

	if tType != nil {
		// Filter by type
		q = q.Where(&Token{Type: *tType})
	}

	err = q.Find(tt).Error

	if err != nil {
		return nil, err
	}

	return tt, nil
}

func (s *GormStore) GetById(id uint64) (*Token, error) {
	var token Token
	err := s.db.First(&token, id).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (s *GormStore) GetByName(name string) (*Token, error) {
	var token Token
	err := s.db.Where("UPPER(name) LIKE UPPER(@name)", sql.Named("name", name)).First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (s *GormStore) Remove(id uint64) error {
	return s.db.Delete(&Token{}, id).Error
}
