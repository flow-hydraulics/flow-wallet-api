package keys

import "github.com/eqlabs/flow-wallet-service/data"

// KeyStore is the interface required by key manager for data storage.
type KeyStore interface {
	AccountKey(address string) (data.Key, error)
}
