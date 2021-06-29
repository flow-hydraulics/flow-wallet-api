# Flow Wallet API

> :warning: This software is a work in progress and is not yet intended for production use.

The Flow Wallet API is a REST HTTP service that allows a developer to integrate wallet functionality into a larger Flow application infrastructure. 
This service can be used by an application that needs to manage Flow user accounts and the assets inside them.

## Features

- Create new Flow accounts
- Securely store account private keys 
- Send a transaction from an account
- Transfer fungible tokens (e.g. FLOW, FUSD)
- Detect fungible token deposits
- _Transfer NFTs (e.g. FLOW, FUSD) (coming soon)_
- _Detect NFT deposits (coming soon)_

View full list of functionality in the [API specification](API.md).

## Background

Some application developers may wish to manage Flow accounts in a fully-custodial fashion,
but without taking on the complexity of building an account management system.

An application may need to support custody of fungible tokens (FLOW, FUSD), non-fungible tokens, or both.

For security and/or legal reasons, 
some developers need to use a custody service running on-premises as part of their existing infrastructure,
rather than a hosted 3rd-party solution.

### Example use cases

- **Custodial NFT Dapp** — an NFT dapp where each user receives a Flow account that is fully managed by the dapp admin. This application requires that each user account can store and transfer NFTs, but does not need to support fungible token custody.
- **FLOW/FUSD Hot Wallet** — an application that allows users to convert fiat currency to FLOW or FUSD. A single admin account would be used as a hot wallet for outgoing payments, and additional deposit accounts would be created to accept incoming payments.
- **Exchange** — a cryptocurrency exchange that is listing FLOW and/or FUSD. Similar to the case above, one or more admin accounts may be used as a hot wallet for outgoing payments, and additional deposit accounts would be created to accept incoming payments.
- **Web Wallet** — a user-facing wallet application that is compatible with Flow dapps. Each user account would be created and managed by the wallet service.

## Installation

The Wallet API is provided as a Docker image:

```sh
docker pull gcr.io/flow-container-registry/wallet-api:v0.0.4
```

### Basic example usage

> This setup requires [Docker](https://docs.docker.com/engine/install/) and the [Flow CLI](https://docs.onflow.org/flow-cli/install/).

Create a configuration file:

```sh
cp .env.example .env
```

Start the Wallet API, Flow Emulator and Postgres:

```sh
docker-compose up -d
```

Deploy the FUSD contract to the emulator:

```sh
flow project deploy -n emulator
```

You can now access the API at http://localhost:3000.

Next, see the [FUSD sample app](/examples/nextjs-fusd-provider)
for an example of how to use this configuration as part of
a complete application.

Once you're finished, run this to stop the containers:

```sh
docker-compose down
```

## Configuration

### Enabled fungible tokens

A comma separated list of fungible tokens and their corresponding addresses enabled for this instance. Make sure to name each token exactly as it is in the corresponding cadence code (FlowToken, FUSD etc.). Include at least FlowToken as functionality without it is undetermined.

Examples:

```sh
ENABLED_TOKENS=FlowToken:0x0ae53cb6e3f42a79
ENABLED_TOKENS=FlowToken:0x0ae53cb6e3f42a79,FUSD:0xf8d6e0586b0a20c7
```

### Database

| Config variable | Environment variable | Description                                                                                      | Default     | Examples                  |
| --------------- | :------------------- | ------------------------------------------------------------------------------------------------ | ----------- | ------------------------- |
| DatabaseType    | `DATABASE_TYPE`      | Type of database driver                                                                          | `sqlite`    | `sqlite`, `psql`, `mysql` |
| DatabaseDSN     | `DATABASE_DSN`       | Data source name ([DSN](https://en.wikipedia.org/wiki/Data_source_name)) for database connection | `wallet.db` | See below                 |

Examples of Database DSN

    mysql://john:pass@localhost:3306/my_db

    postgresql://postgres:postgres@localhost:5432/postgres

    user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local

    host=localhost user=gorm password=gorm dbname=gorm port=9920 sslmode=disable TimeZone=Asia/Shanghai

For more: https://gorm.io/docs/connecting_to_the_database.html

### Google KMS setup

Note: In order to use Google KMS for remote key management you'll need a Google Cloud Platform account.

Pre-requisites:

1. Create a new Project if you don't have one already. You'll need the Project ID later.
2. Enable Cloud Key Management Service (KMS) API for the project, Security -> [Cryptographic Keys](https://console.cloud.google.com/security/kms).
3. Create a new Key Ring for your wallet (or use an existing Key Ring), Security -> Cryptographic Keys -> [Create Key Ring](https://console.cloud.google.com/security/kms/keyring/create), you'll need the Location ID (or _Location_) and Key Ring ID (or _Name_) later.

Using a Service Account to access the KMS API (see [official docs](https://cloud.google.com/docs/authentication/getting-started) for more);

1. Create a new Service Account, IAM & Admin -> Service Accounts -> [Create Service Account](https://console.cloud.google.com/iam-admin/serviceaccounts/create)
2. Use the roles `Cloud KMS Admin` & `Cloud KMS Signer/Verifier` or grant the required permissions through a custom role (NOTE: deletion not supported yet):
   - `cloudkms.cryptoKeyVersions.useToSign`
   - `cloudkms.cryptoKeyVersions.viewPublicKey`
   - `cloudkms.cryptoKeys.create`
3. After creating the Service Account, select Manage Keys from the Actions menu in the Service Account listing.
4. Create a new key, Add Key -> Create New key, and select JSON as the key type.
5. Save the JSON file.

Configure the Google KMS client library by setting the environment variable `GOOGLE_APPLICATION_CREDENTIALS`:

```
export GOOGLE_APPLICATION_CREDENTIALS="/home/example/path/to/service-account-file.json"
```

Configure Google KMS as the key storage for `flow-wallet-service` and set the necessary environment variables:

| Config variable | Environment variable     | Description      | Default | Examples                    |
| --------------- | ------------------------ | ---------------- | ------- | --------------------------- |
| DefaultKeyType  | `DEFAULT_KEY_TYPE`       | Default key type | `local` | `local`, `google_kms`       |
| ProjectID       | `GOOGLE_KMS_PROJECT_ID`  | GCP Project ID   | -       | `flow-wallet-example`       |
| LocationID      | `GOOGLE_KMS_LOCATION_ID` | GCP Location ID  | -       | `europe-north1`, `us-west1` |
| KeyRingID       | `GOOGLE_KMS_KEYRING_ID`  | GCP Key Ring ID  | -       | `example-wallet-keyring`    |

### All possible environment variables

```
HOST=
PORT=3000
ACCESS_API_HOST=localhost:3569

ENABLED_TOKENS=FlowToken:0x0ae53cb6e3f42a79

DATABASE_DSN=wallet.db
DATABASE_TYPE=sqlite

ADMIN_ADDRESS=
ADMIN_KEY_INDEX=0
ADMIN_KEY_TYPE=local
ADMIN_PRIVATE_KEY=
CHAIN_ID=flow-emulator
DEFAULT_KEY_TYPE=local
DEFAULT_KEY_INDEX=0
DEFAULT_KEY_WEIGHT=-1
DEFAULT_SIGN_ALGO=ECDSA_P256
DEFAULT_HASH_ALGO=SHA3_256
ENCRYPTION_KEY=

GOOGLE_APPLICATION_CREDENTIALS=
GOOGLE_KMS_PROJECT_ID=
GOOGLE_KMS_LOCATION_ID=
```

## API Specification

[View the full Wallet API specification](API.md).

## Credit

The Flow Wallet API is developed and maintained by [Equilibrium](https://equilibrium.co/),
with support from the Flow core contributors.

<a href="https://equilibrium.co/"><img src="equilibrium.svg" alt="Equilibrium" width="200"/></a>
