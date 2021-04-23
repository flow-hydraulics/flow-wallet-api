package postgres

type KeyStore struct{}

func NewKeyStore() (*KeyStore, error) {
	return &KeyStore{}, nil
}
