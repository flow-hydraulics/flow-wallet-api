# Flow NFT Wallet service

A custodial wallet service for NTFs on the Flow blockchain.

## Dev

Dependencies:

- docker
- go

Run:

    ./scripts/emulator.sh
    cp .env.example .env
    # edit .env
    go run main.go

## Configuration

### Database

| Config variable |   ENV   | descrpition                                                                                      | default   | examples                                                                                                                                                                                                                                                                                                      |
| --------------- | :-----: | ------------------------------------------------------------------------------------------------ | --------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| DatabaseType    | DB_TYPE | Type of database driver                                                                          | sqlite    |                                                                                                                                                                                                                                                                                                               |
| DatabaseDSN     | DB_DSN  | Data source name ([DSN](https://en.wikipedia.org/wiki/Data_source_name)) for database connection | wallet.db | mysql://john:pass@localhost:3306/my_db<br><br>user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local<br><br>host=localhost user=gorm password=gorm dbname=gorm port=9920 sslmode=disable TimeZone=Asia/Shanghai<br><br> For more: https://gorm.io/docs/connecting_to_the_database.html |
