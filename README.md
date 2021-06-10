# Flow Wallet API Demo (Node.js + express)

> :warning: This demo is a work in progress.


This is a demonstration of a RESTful API that
implements a simple custodial wallet service for the Flow blockchain.

## Functionality

### 1. Admin

- [x] Single admin account (hot wallet)
- [x] [Create user accounts (using admin account)](https://github.com/onflow/flow-wallet-api-node-demo/issues/1)

### 2. Transaction Execution

- [x] Send an arbitrary transaction from the admin account
- [x] Send an arbitrary transaction from a user account

### 3. Fungible Tokens

- [x] Send fungible token withdrawals from admin account (FLOW, FUSD)
- [ ] [Detect fungible token deposits to admin account (FLOW, FUSD)](https://github.com/onflow/flow-wallet-api-node-demo/issues/2)
- [x] [Send fungible token withdrawals from a user account (FLOW, FUSD)](https://github.com/onflow/flow-wallet-api-node-demo/issues/3)
- [ ] [Detect fungible token deposits to a user account (FLOW, FUSD)](https://github.com/onflow/flow-wallet-api-node-demo/issues/4)
- [ ] View the fungible token balance of the admin account
- [ ] View the fungible token balance of a user account

### 4. Non-Fungible Tokens

- [ ] Set up admin account with non-fungible token collections (`NFT.Collection`)
- [ ] Send non-fungible token withdrawals from admin account
- [ ] Detect non-fungible token deposits to admin account
- [ ] Set up a user account with non-fungible token collections (`NFT.Collection`)
- [ ] Send non-fungible token withdrawals from a user account
- [ ] Detect non-fungible token deposits to a user account
- [ ] View the non-fungible tokens owned by the admin account
- [ ] View the non-fungible tokens owned by a user account

## Local Development

> This local development environment uses the 
> [Flow Emulator](https://docs.onflow.org/emulator) to 
> simulate the real Flow network.

### Install the Flow CLI

First, install the [Flow CLI](https://docs.onflow.org/flow-cli/install/).

### Install dependencies and configure environment

```sh
npm install

cp .env.example .env
```

### Start the database and emulator

Use Docker Compose to launch Postgres and the [Flow Emulator](https://docs.onflow.org/emulator):

```sh
npm run docker-local-network
```

### Start the server

```sh
npm run dev
```

## Deploy with Docker

To deploy this API as a Docker container in your infrastructure,
either build from source or use the pre-built image:

```sh
docker pull gcr.io/flow-container-registry/flow-wallet-api-demo:latest
```

The Docker Compose sample configurations
in this repository show how to configure this application when
running as a Docker container.

### Emulator (Local Development)

> This example shows how to connect the Docker container
> to an instance of the [Flow Emulator](https://docs.onflow.org/emulator).

Configuration: [docker-compose.emulator.yml](docker-compose.emulator.yml)

```sh
cp .env.emulator.example .env

docker-compose -f docker-compose.emulator.yml up
```

Once the emulator is running, 
you will need to deploy the FUSD contract:

```sh
npm run dev-deploy-contracts
```

### Testnet

> This example shows how to connect the Docker container
> to Flow Testnet.

First you'll need a Testnet account. Here's how to make one:

#### Generate a key pair 

Generate a new key pair with the Flow CLI:

```sh
flow keys generate
```

_⚠️ Make sure to save these keys in a safe place, you'll need them later._

#### Create your account

Go to the [Flow Testnet Faucet](https://testnet-faucet.onflow.org/) to create a new account. Use the **public key** from the previous step.

#### Save your keys

After your account has been created, save the address and private key in the `.env` file:

```sh
cp .env.testnet.example .env
```

```sh
# Replace these values with your own!
FLOW_ADDRESS=0xabcdef12345689
FLOW_PRIVATE_KEY=aaaaaa...aaaaaa
```

### Start the Docker containers

Configuration: [docker-compose.testnet.yml](docker-compose.testnet.yml)

```sh
docker-compose -f docker-compose.testnet.yml up
```

## API Routes

### Accounts

#### List all accounts

`GET /v1/accounts`

Example

```sh
curl --request GET \
  --url http://localhost:3000/v1/accounts
```

```json
[
  {
    "address": "0xf8d6e0586b0a20c7"
  },
  {
    "address": "0xe467b9dd11fa00df"
  }
]
```

---

#### Get an account

`GET /v1/accounts/{address}`

Parameters

- `address`: The address of the account (e.g. "0xf8d6e0586b0a20c7")

Example

```sh
curl --request GET \
  --url http://localhost:3000/v1/accounts/0xf8d6e0586b0a20c7
```

```json
{
  "address": "0xf8d6e0586b0a20c7"
}
```

---

#### Create an account

`POST /v1/accounts`

Example

```sh
curl --request POST \
  --url http://localhost:3000/v1/accounts
```

```json
{
  "address": "0xe467b9dd11fa00df"
}
```

---

### Transaction Execution

#### Execute a transaction

> :warning: Not yet implemented

`POST /v1/accounts/{address}/transactions`

Parameters

- `address`: The address of the account (e.g. "0xf8d6e0586b0a20c7")

Body (JSON)

- `code`: The Cadence code to execute in the transaction
  - The code must always specify exactly one authorizer (i.e. `prepare(auth: AuthAccount)`)

Example

```sh
curl --request POST \
  --url http://localhost:3000/v1/accounts/0xf8d6e0586b0a20c7/transactions \
  --header 'Content-Type: application/json' \
  --data '{ "code": "transaction { prepare(auth: AuthAccount) { log(\"Hello, World!\") } }" }'
```

```json
{
  "transactionId": "18647b584a03345f3b2d2c4d9ab2c4179ae1b124a7f62ef9f33910e5ca8b353c",
  "error": null,
}
```

---

### Fungible Tokens

Supported tokens:
- `FLOW`
- `FUSD`

#### List all tokens

`GET /v1/accounts/{address}/fungible-tokens`

Parameters

- `address`: The address of the account (e.g. "0xf8d6e0586b0a20c7")

Example

```sh
curl --request GET \
  --url http://localhost:3000/v1/accounts/0xf8d6e0586b0a20c7/fungible-tokens
```

```json
[
  {
    "name": "flow"
  },
  {
    "name": "fusd"
  }
]
```

---

#### Get details of a token type

`GET /v1/accounts/{address}/fungible-tokens/{tokenName}`

Parameters

- `address`: The address of the account (e.g. "0xf8d6e0586b0a20c7")
- `tokenName`: The name of the fungible token (e.g. "flow")

Example

```sh
curl --request GET \
  --url http://localhost:3000/v1/accounts/0xf8d6e0586b0a20c7/fungible-tokens/flow
```

```json
{
  "name": "flow", 
  "balance": "42.0"
}
```

---

#### List all withdrawals of a token type

> :warning: Not yet implemented

`GET /v1/accounts/{address}/fungible-tokens/{tokenName}/withdrawals`

Parameters

- `address`: The address of the account (e.g. "0xf8d6e0586b0a20c7")
- `tokenName`: The name of the fungible token (e.g. "flow")

---

#### Get details of a token withdrawal

> :warning: Not yet implemented

`GET /v1/accounts/{address}/fungible-tokens/{tokenName}/withdrawals/{transactionId}`

Parameters

- `address`: The address of the account (e.g. "0xf8d6e0586b0a20c7")
- `tokenName`: The name of the fungible token (e.g. "flow")
- `transactionId`: The Flow transaction ID for the withdrawal

---

#### Create a token withdrawal

`POST /v1/accounts/{address}/fungible-tokens/{tokenName}/withdrawals`

Parameters

- `address`: The address of the account (e.g. "0xf8d6e0586b0a20c7")
- `tokenName`: The name of the fungible token (e.g. "flow")

Body (JSON)

- `amount`: The number of tokens to transfer (e.g. "123.456")
  - Must be a fixed-point number with a maximum of 8 decimal places
- `recipient`: The Flow address of the recipient (e.g. "0xf8d6e0586b0a20c7")

Example

```sh
curl --request POST \
  --url http://localhost:3000/v1/accounts/0xf8d6e0586b0a20c7/fungible-tokens/fusd/withdrawls \
  --header 'Content-Type: application/json' \
  --data '{ "recipient": "0xe467b9dd11fa00df", "amount": "123.456" }'
```

```json
{
  "transactionId": "18647b584a03345f3b2d2c4d9ab2c4179ae1b124a7f62ef9f33910e5ca8b353c",
  "recipient": "0xe467b9dd11fa00df",
  "amount": "123.456"
}
```

---

### Non-Fungible Tokens

> :warning: Not yet implemented

#### List all tokens

> :warning: Not yet implemented

`GET /v1/accounts/{address}/non-fungible-tokens`

Parameters

- `address`: The address of the account (e.g. "0xf8d6e0586b0a20c7")

Example

```sh
curl --request GET \
  --url http://localhost:3000/v1/accounts/0xf8d6e0586b0a20c7/non-fungible-tokens
```

---

#### Get details of a token

> :warning: Not yet implemented

`GET /v1/accounts/{address}/non-fungible-tokens/{tokenName}`

Parameters

- `address`: The address of the account (e.g. "0xf8d6e0586b0a20c7")
- `tokenName`: The name of the non-fungible token (e.g. "nba-top-shot-moment")

---

#### List all withdrawals of a token type

> :warning: Not yet implemented

`GET /v1/accounts/{address}/non-fungible-tokens/{tokenName}/withdrawals`

Parameters

- `address`: The address of the account (e.g. "0xf8d6e0586b0a20c7")
- `tokenName`: The name of the non-fungible token (e.g. "nba-top-shot-moment")

---

#### Get details of a token withdrawal

> :warning: Not yet implemented

`GET /v1/accounts/{address}/non-fungible-tokens/{tokenName}/withdrawals/{transactionId}`

Parameters

- `address`: The address of the account (e.g. "0xf8d6e0586b0a20c7")
- `tokenName`: The name of the non-fungible token (e.g. "nba-top-shot-moment")
- `transactionId`: The Flow transaction ID for the withdrawal

---

#### Create a token withdrawal

> :warning: Not yet implemented

`POST /v1/accounts/{address}/non-fungible-tokens/{tokenName}/withdrawals`

Parameters

- `address`: The address of the account (e.g. "0xf8d6e0586b0a20c7")
- `tokenName`: The name of the non-fungible token (e.g. "nba-top-shot-moment")

Body (JSON)

- `recipient`: The Flow address of the recipient (e.g. "0xf8d6e0586b0a20c7")
