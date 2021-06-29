package templates

// Store manages data regarding templates.
type Store interface {
	Insert(*Token) error
	List() (*[]Token, error)
	Get(id uint64) (*Token, error)
	Remove(id uint64) error
}
