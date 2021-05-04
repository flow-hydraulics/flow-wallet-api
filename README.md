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

## Test

Run:

    ./scripts/emulator.sh
    cp .env.example .env.test
    # edit .env.test
    go test -v ./...

## Configuration

### Database

| Config variable | ENV       | Description                                                                                      | Default   | Examples  |
| --------------- | :-------- | ------------------------------------------------------------------------------------------------ | --------- | --------- |
| DatabaseType    | `DB_TYPE` | Type of database driver                                                                          | sqlite    |           |
| DatabaseDSN     | `DB_DSN`  | Data source name ([DSN](https://en.wikipedia.org/wiki/Data_source_name)) for database connection | wallet.db | See below |

Examples of Database DSN

    mysql://john:pass@localhost:3306/my_db

    user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local

    host=localhost user=gorm password=gorm dbname=gorm port=9920 sslmode=disable TimeZone=Asia/Shanghai

For more: https://gorm.io/docs/connecting_to_the_database.html

### Google KMS Setup

Note: In order to use Google KMS for remote key management you'll need a Google Cloud Platform account.

Pre-requisites:

1. Create a new Project if you don't have one already, you'll need the Project ID later.
2. Enable Cloud Key management Service (KMS) API for the project, Security -> [Cryptographic Keys](https://console.cloud.google.com/security/kms)
3. Create a new Key Ring for your wallet (or use an existing Key Ring), Security -> Cryptographic Keys -> [Create Key Ring](https://console.cloud.google.com/security/kms/keyring/create), you'll need the Location ID (or _Location_) and Key Ring ID (or _Name_) later.

Using a Service Account to access the KMS API (see [official docs](https://cloud.google.com/docs/authentication/getting-started) for more);

1. Create a new Service Account, IAM & Admin -> Service Accounts -> [Create Service Account](https://console.cloud.google.com/iam-admin/serviceaccounts/create)
2. Grant the required permissions through a custom Role or use the `Cloud KMS Admin` role.
3. After creating the Service Account, select Manage Keys from the Actions menu in the Service Account listing.
4. Create a new key, Add Key -> Create new key, and select JSON as the key type
5. Save the JSON file

Configure the Google KMS client library by setting the environment variable `GOOGLE_APPLICATION_CREDENTIALS`;

```
export GOOGLE_APPLICATION_CREDENTIALS="/home/example/path/to/service-account-file.json"
```

Configure Google KMS as the key storage for `flow-wallet-service` and set the necessary environment variables;

| Config variable   | Environment variable     | Description         | Default | Examples                    |
| ----------------- | ------------------------ | ------------------- | ------- | --------------------------- |
| DefaultKeyStorage | `DEFAULT_KEY_STORAGE`    | Default key storage | `local` | `local`, `google_kms`       |
| ProjectID         | `GOOGLE_KMS_PROJECT_ID`  | GCP Project ID      | -       | `flow-wallet-example`       |
| LocationID        | `GOOGLE_KMS_LOCATION_ID` | GCP Location ID     | -       | `europe-north1`, `us-west1` |
| KeyRingID         | `GOOGLE_KMS_KEYRING_ID`  | GCP Key Ring ID     | -       | `example-wallet-keyring`    |
