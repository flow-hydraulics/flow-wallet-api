# Flow Wallet API Demo (Node.js + express)

> :warning: This demo is a work in progress.


This is a demonstration of a RESTful API that
implements a simple custodial wallet service for the Flow blockchain.

## Functionality

### 1. Admin

- [x] Single admin account (hot wallet)
- [ ] [Create user accounts (using admin account)](https://github.com/onflow/flow-wallet-api-node-demo/issues/1)

### 2. Fungible Tokens

- [x] Send fungible token withdrawals from admin account (FLOW, FUSD)
- [ ] [Detect fungible token deposits to admin account (FLOW, FUSD)](https://github.com/onflow/flow-wallet-api-node-demo/issues/2)
- [ ] [Send fungible token withdrawals from a user account (FLOW, FUSD)](https://github.com/onflow/flow-wallet-api-node-demo/issues/3)
- [ ] [Detect fungible token deposits to a user account (FLOW, FUSD)](https://github.com/onflow/flow-wallet-api-node-demo/issues/4)
- [ ] View the fungible token balance of the admin account
- [ ] View the fungible token balance of a user account

### 3. Non-Fungible Tokens

- [ ] Set up admin account with non-fungible token collections (`NFT.Collection`)
- [ ] Send non-fungible token withdrawals from admin account
- [ ] Detect non-fungible token deposits to admin account
- [ ] Set up a user account with non-fungible token collections (`NFT.Collection`)
- [ ] Send non-fungible token withdrawals from a user account
- [ ] Detect non-fungible token deposits to a user account
- [ ] View the non-fungible tokens owned by the admin account
- [ ] View the non-fungible tokens owned by a user account

## Getting Started

### Install the Flow CLI

First, install the [Flow CLI](https://docs.onflow.org/flow-cli/install/).

### Install dependencies and configure environment

```sh
npm install

cp env.example .env
```

### Start the database and emulator

Use Docker Compose to launch Postgres and the [Flow Emulator](https://docs.onflow.org/emulator):

```sh
docker-compose up -d
```

### Start the server

```sh
npm run start
```

## API Routes

### Accounts

#### List all accounts

> :warning: Not yet implemented

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

### Get an account

> :warning: Not yet implemented

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

### Create an account

> :warning: Not yet implemented

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

`GET /v1/accounts/{v1/accounts/{address}/fungible-tokens}/fungible-tokens/{tokenName}/withdrawals`

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
