package store

import "github.com/google/uuid"

type Account struct {
	ID      uuid.UUID `db:"id" json:"-"`
	Address string    `db:"address" json:"address"`
}

type Transaction struct {
	ID   uuid.UUID `db:"id" json:"-"`
	TxId string    `db:"txId" json:"txId"`
}
