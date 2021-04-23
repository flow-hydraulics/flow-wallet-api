package handlers

import (
	"log"

	"github.com/eqlabs/flow-nft-wallet-service/pkg/store"
	"github.com/onflow/flow-go-sdk/client"
)

type Server struct {
	l  *log.Logger
	fc *client.Client
	db *store.DataStore
	ks *store.KeyStore
}

func New(
	l *log.Logger,
	fc *client.Client,
	db *store.DataStore,
	ks *store.KeyStore) *Server {
	return &Server{l, fc, db, ks}
}
