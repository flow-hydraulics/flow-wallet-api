# Flow Wallet API

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

View full list of functionality in the [API documentation](https://flow-hydraulics.github.io/flow-wallet-api/).

## Background

Some application developers may wish to manage Flow accounts in a fully-custodial fashion,
but without taking on the complexity of building an account management system.

An application may need to support custody of fungible tokens (FLOW, FUSD), non-fungible tokens, or both.

For security and/or legal reasons,
some developers need to use a custody service running on-premises as part of their existing infrastructure,
rather than a hosted 3rd-party solution.

### Example use cases

- **FLOW/FUSD Hot Wallet** — an application that allows users to convert fiat currency to FLOW or FUSD. A single admin account would be used as a hot wallet for outgoing payments, and additional deposit accounts would be created to accept incoming payments.
- **Exchange** — a cryptocurrency exchange that is listing FLOW and/or FUSD. Similar to the case above, one or more admin accounts may be used as a hot wallet for outgoing payments, and additional deposit accounts would be created to accept incoming payments.
- **Web Wallet** — a user-facing wallet application that is compatible with Flow dapps. Each user account would be created and managed by the wallet service.

## API Specification

View the [Wallet API documentation and OpenAPI (Swagger) specification](https://flow-hydraulics.github.io/flow-wallet-api/).

## Installation

The Wallet API is provided as a Docker image:

```sh
docker pull ghcr.io/flow-hydraulics/flow-wallet-api:latest
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

You can now access the API at http://localhost:3000/v1/accounts.

Next, see the [FUSD sample app](/examples/nextjs-fusd-provider)
for an example of how to use this configuration as part of
a complete application.

Once you're finished, run this to stop the containers:

```sh
docker-compose down
```

## Configuration

### Enabled fungible tokens

A comma separated list of _fungible tokens_ and their corresponding addresses enabled for this instance. Make sure to name each token exactly as it is in the corresponding cadence code (FlowToken, FUSD etc.). Include at least FlowToken as functionality without it is undetermined.

**NOTE:** It is necessary to add a 3rd parameter "lowercamelcase" name for each token. For FlowToken this would be "flowToken" and for FUSD "fusd". This is used to construct the vault name, receiver name and balance name in generic transaction templates. Consult the contract code for each token to derive the proper name (search for `.*Vault`, `.*Receiver`, `.*Balance`)

**NOTE:** Non-fungible tokens can _not_ be enabled using environment variables. Use the API endpoints for that.

Examples:

    FLOW_WALLET_ENABLED_TOKENS=FlowToken:0x0ae53cb6e3f42a79:flowToken,FUSD:0xf8d6e0586b0a20c7:fusd

### Database

| Config variable | Environment variable        | Description                                                                                      | Default     | Examples                  |
| --------------- | :-------------------------- | ------------------------------------------------------------------------------------------------ | ----------- | ------------------------- |
| DatabaseType    | `FLOW_WALLET_DATABASE_TYPE` | Type of database driver                                                                          | `sqlite`    | `sqlite`, `psql`, `mysql` |
| DatabaseDSN     | `FLOW_WALLET_DATABASE_DSN`  | Data source name ([DSN](https://en.wikipedia.org/wiki/Data_source_name)) for database connection | `wallet.db` | See below                 |

Examples of Database DSN

    mysql://john:pass@localhost:3306/my_db

    postgresql://postgres:postgres@localhost:5432/postgres

    user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local

    host=localhost user=gorm password=gorm dbname=gorm port=9920 sslmode=disable TimeZone=Asia/Shanghai

For more: https://gorm.io/docs/connecting_to_the_database.html

To learn more about database schema versioning and migrations, read [MIGRATIONS.md](MIGRATIONS.MD).

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

Configure Google KMS as the key storage for `flow-wallet-api` and set the necessary environment variables:

| Config variable | Environment variable                 | Description      | Default | Examples                    |
| --------------- | ------------------------------------ | ---------------- | ------- | --------------------------- |
| DefaultKeyType  | `FLOW_WALLET_DEFAULT_KEY_TYPE`       | Default key type | `local` | `local`, `google_kms`       |
| ProjectID       | `FLOW_WALLET_GOOGLE_KMS_PROJECT_ID`  | GCP Project ID   | -       | `flow-wallet-example`       |
| LocationID      | `FLOW_WALLET_GOOGLE_KMS_LOCATION_ID` | GCP Location ID  | -       | `europe-north1`, `us-west1` |
| KeyRingID       | `FLOW_WALLET_GOOGLE_KMS_KEYRING_ID`  | GCP Key Ring ID  | -       | `example-wallet-keyring`    |

### Google KMS for admin account

If you want to use a key stored in Google KMS for the admin account, just pass the resource identifier as the private key (`FLOW_WALLET_ADMIN_PRIVATE_KEY`) and set `FLOW_WALLET_ADMIN_KEY_TYPE` to `google_kms`.

**Creating an account on testnet via the faucet:**

1. When generating the key, choose "Asymmetric sign" as the purpose and "Elliptic Curve P-256 - SHA256 Digest" as the key type and algorithm (other combinations may work but these have been confirmed to work)
2. Download the public key from Google KMS in PEM format (should have a `.pub` ending)
3. Run it through `flow keys decode pem --from-file <filename>`
4. Copy the "Public Key" part
5. Go to https://testnet-faucet-v2.onflow.org
6. Paste the copied public key in the form
7. **IMPORTANT**: Choose **SHA2_256** as the hash algorithm (_SHA3_256_ won't work with the key setup above)
8. Copy the new address and use it as the `FLOW_WALLET_ADMIN_ADDRESS`
9. Set `FLOW_WALLET_ADMIN_PRIVATE_KEY` to the resource id of the key
10. Set `FLOW_WALLET_ADMIN_KEY_TYPE` to `google_kms`

Example environment:

    FLOW_WALLET_ADMIN_ADDRESS=0x1234567890123456
    FLOW_WALLET_ADMIN_PRIVATE_KEY=projects/<project_id>/locations/<location_id>/keyRings/<keyring_id>/cryptoKeys/<key_name>/cryptoKeyVersions/<version_number> # Make sure this ends with the version number
    FLOW_WALLET_ADMIN_KEY_TYPE=google_kms

NOTE: This will mess up the docker-compose setup (emulator won't start) as it uses `FLOW_WALLET_ADMIN_PRIVATE_KEY` as `FLOW_SERVICEPRIVATEKEY`. It will cause an encoding error on the emulator.


### AWS KMS setup

Note: In order to use AWS KMS for remote key management you'll need an AWS account.
Note: Custom key stores are not supported.

Pre-requisites:

1. AWS credentials for an account that has access to KMS

Configure AWS KMS as the key storage for `flow-wallet-api` and set the necessary environment variables, with the default key type as `aws_kms`:

| Config variable | Environment variable                 | Description            | Default | Examples                         |
| --------------- | ------------------------------------ | ---------------------- | ------- | -------------------------------- |
| DefaultKeyType  | `FLOW_WALLET_DEFAULT_KEY_TYPE`       | Default key type       | `local` | `aws_kms`, `google_kms`, `local` |
| -               | `AWS_REGION`                         | AWS KMS Region         | -       | `eu-central-1`, `us-west-1`      |
| -               | `AWS_ACCESS_KEY_ID`                  | AWS access key ID      | -       | `AKIAXXX123FOOBAR1234`           |
| -               | `AWS_SECRET_ACCESS_KEY`              | AWS secret access key  | -       | `FooBaRBaZ12345...`              |

### AWS KMS for admin account

If you want to use a key stored in AWS KMS for the admin account, just pass the ARN (e.g. `arn:aws:kms:eu-central-1:012345678910:key/00000000-aaaa-bbbb-cccc-12345678910`) of the admin key as the private key (`FLOW_WALLET_ADMIN_PRIVATE_KEY`) and set `FLOW_WALLET_ADMIN_KEY_TYPE` to `aws_kms`.

When testing make sure to add the key to the admin account. You can convert the AWS public key (e.g. `aws.pem`) you downloaded/copied from AWS with flow-cli;

```
flow keys decode pem --from-file=aws.pem --sig-algo "ECDSA_secp256k1"
```

### All possible configuration variables

Refer to [configs/configs.go](configs/configs.go) for details and documentation.

## Credit

The Flow Wallet API is developed and maintained by [Equilibrium](https://equilibrium.co/),
with support from the Flow core contributors.

<a href="https://equilibrium.co/"><img src="equilibrium.svg" alt="Equilibrium" width="200"/></a>
