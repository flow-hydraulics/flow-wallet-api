# Database Schema Migrations

## Introduction

`flow-wallet-api` supports currently three database engines:
- `sqlite`
- `mysql`
- `postgresql`

[GORM](https://gorm.io/) is used as an ORM for simplifying database operations.
It provides support for [automatically
migrating](https://gorm.io/docs/migration.html) DB schema migrations, but no
versioning. To address schema versioning,
[gormigrate](https://github.com/go-gormigrate/gormigrate) is used.

## Structure for DB migrations

Each migration must be created as a separate package under
`migrations/internal`. The current pattern for package name is `m + <YYYYMMDD`
- e.g. `m20210922`.

In order to activate / register the migration, it must be added to migration
list in `migrations/migrations.go`.


`ID` of the migration can be free form string, but the date from migration
package is generally preferred (i.e. `20210922`).

## Migrating database to a specific schema version

To migrate to a specific schema version, the `ID` of the migration can be
configured to `FLOW_WALLET_DATABASE_VERSION` environment variable.

**NOTE:** Rollback to old schema version **will** result in data loss if added
columns / tables had been populated before rollback.


When the database version configuration flag has been left empty, the database
is migrated to latest version.

## Database migrations in clustered environments

When there are more than one instances of `flow-wallet-api` running, the DB
migrations must be performed under a controlled process, depending on changes.

There's no distributed locking mechanism implemented because the deployment
environment can vary, but depending on changes it wouldn't prevent
inter-version compatibility changes in case the schema/code change introduces
one.
