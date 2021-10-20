package migrations

import (
	"github.com/flow-hydraulics/flow-wallet-api/migrations/internal/m20210922"
	"github.com/go-gormigrate/gormigrate/v2"
)

func List() []*gormigrate.Migration {
	ms := []*gormigrate.Migration{
		&gormigrate.Migration{
			ID:       m20210922.ID,
			Migrate:  m20210922.Migrate,
			Rollback: m20210922.Rollback,
		},
	}
	return ms
}
