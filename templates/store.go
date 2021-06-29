package templates

// Store manages data regarding templates.
type Store interface {
	Insert(*Token) error
	List(*TokenType) (*[]Token, error)
	GetById(id uint64) (*Token, error)
	GetByName(name string) (*Token, error)
	Remove(id uint64) error
}
