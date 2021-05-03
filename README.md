# Flow Wallet service

A custodial wallet service for tokens on the Flow blockchain.

## Dev

Dependencies:

- docker
- go

Run:

    ./scripts/emulator.sh
    cp .env.example .env
    # edit .env
    go run main.go

_Note:
The emulator creates new account addresses deterministically. This means that deleting the emulators docker volume will cause the emulator to start from the beginning and give the same addresses as before possibly ending in duplicate key errors in database._

## Configuration

### Database

| Config variable |   ENV   | descrpition                                                                                      | default   | examples  |
| --------------- | :-----: | ------------------------------------------------------------------------------------------------ | --------- | --------- |
| DatabaseType    | DB_TYPE | Type of database driver                                                                          | sqlite    |           |
| DatabaseDSN     | DB_DSN  | Data source name ([DSN](https://en.wikipedia.org/wiki/Data_source_name)) for database connection | wallet.db | See below |

Examples of Database DSN

    mysql://john:pass@localhost:3306/my_db

    user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local

    host=localhost user=gorm password=gorm dbname=gorm port=9920 sslmode=disable TimeZone=Asia/Shanghai

For more: https://gorm.io/docs/connecting_to_the_database.html
