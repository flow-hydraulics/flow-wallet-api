package templates

// Store manages data regarding templates.
type Store interface {
	Insert(*Token) error
	List(TokenType) ([]BasicToken, error)
	ListFull(TokenType) ([]Token, error)
	GetById(id uint64) (*Token, error)
	GetByName(name string) (*Token, error)
	Remove(id uint64) error
	// Insert a token that is available only for this instances runtime (in-memory)
	// Used when enabling a token via environment variables
	InsertTemp(*Token)
}
