package templates

import (
	"database/sql"
	"strings"

	"gorm.io/gorm"
)

type GormStore struct {
	db        *gorm.DB
	tempStore map[string]*Token
}

func NewGormStore(db *gorm.DB) *GormStore {
	return &GormStore{db, make(map[string]*Token)}
}

func (s *GormStore) Insert(q *Token) error {
	return s.db.Omit("ID").Create(q).Error
}

func (s *GormStore) List(tType TokenType) (*[]BasicToken, error) {
	var err error

	fromTemp := make([]BasicToken, 0, len(s.tempStore))
	for _, t := range s.tempStore {
		if tType == NotSpecified || t.Type == tType {
			fromTemp = append(fromTemp, t.BasicToken())
		}
	}

	fromDB := []BasicToken{}

	q := s.db.Model(&Token{})

	if tType != NotSpecified {
		// Filter by type
		q = q.Where(&Token{Type: tType})
	}

	err = q.Find(&fromDB).Error

	if err != nil {
		return nil, err
	}

	result := append(fromDB, fromTemp...)

	return &result, nil
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
	fromTemp, ok := s.tempStore[strings.ToLower(name)]
	if ok {
		return fromTemp, nil
	}

	var fromDB Token
	if err := s.db.Where("LOWER(name) LIKE LOWER(@name)", sql.Named("name", name)).First(&fromDB).Error; err != nil {
		return nil, err
	}

	return &fromDB, nil
}

func (s *GormStore) Remove(id uint64) error {
	return s.db.Delete(&Token{}, id).Error
}

func (s *GormStore) InsertTemp(token *Token) {
	s.tempStore[strings.ToLower(token.Name)] = token
}
