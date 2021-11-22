package migrations

import (
	"github.com/flow-hydraulics/flow-wallet-api/migrations/internal/m20210922"
	"github.com/flow-hydraulics/flow-wallet-api/migrations/internal/m20211005"
	"github.com/flow-hydraulics/flow-wallet-api/migrations/internal/m20211015"
	"github.com/flow-hydraulics/flow-wallet-api/migrations/internal/m20211118"
	"github.com/flow-hydraulics/flow-wallet-api/migrations/internal/m20211130"
	"github.com/flow-hydraulics/flow-wallet-api/migrations/internal/m20211202"
	"github.com/go-gormigrate/gormigrate/v2"
)

func List() []*gormigrate.Migration {
	ms := []*gormigrate.Migration{
		{
			ID:       m20210922.ID,
			Migrate:  m20210922.Migrate,
			Rollback: m20210922.Rollback,
		},
		{
			ID:       m20211005.ID,
			Migrate:  m20211005.Migrate,
			Rollback: m20211005.Rollback,
		},
		{
			ID:       m20211015.ID,
			Migrate:  m20211015.Migrate,
			Rollback: m20211015.Rollback,
		},
		{
			ID:       m20211118.ID,
			Migrate:  m20211118.Migrate,
			Rollback: m20211118.Rollback,
		},
		{
			ID:       m20211130.ID,
			Migrate:  m20211130.Migrate,
			Rollback: m20211130.Rollback,
		},
		{
			ID:       m20211202.ID,
			Migrate:  m20211202.Migrate,
			Rollback: m20211202.Rollback,
		},
	}
	return ms
}
