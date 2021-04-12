# Flow Wallet API Demo (Node.js + express)

This is a demonstration of a RESTful API that
implements a simple custodial wallet service for the Flow blockchain.

## Functionality

- [x] Single admin account (hot wallet)
- [x] Send fungible token withdrawals from admin account (FLOW, FUSD)
- [ ] Detect fungible token deposits to admin account (FLOW, FUSD)
- [ ] Create user accounts from admin account
- [ ] Send fungible token withdrawals from a user account (FLOW, FUSD)
- [ ] Detect fungible token deposits to a user account (FLOW, FUSD)
- [ ] Send non-fungible token withdrawals from admin account
- [ ] Detect non-fungible token deposits to admin account
- [ ] Send non-fungible token withdrawals from a user account
- [ ] Detect non-fungible token deposits to aa user account

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

### Deploy token contracts to the emulator

```sh
flow project deploy --network=emulator
```

### Set up the database

```sh
npm run db-migrate-dev
```

### Start the server!

```sh
npm run start
```

## API Routes

### Fungible Tokens

Supported tokens:
- `FLOW`
- `FUSD`

#### List all tokens

`GET /v1/fungible-tokens`

Example

```sh
curl --request GET \
  --url http://localhost:3000/v1/fungible-tokens
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

#### Get details of a token

`GET /v1/fungible-tokens/{tokenName}`

Parameters

- `tokenName`: The name of the fungible token (e.g. "flow")

Example

```sh
curl --request GET \
  --url http://localhost:3000/v1/fungible-tokens/flow
```

```json
{
  "name": "flow"
}
```

---

#### List all withdrawals of a token type

:warning: _Not yet implemented_

`GET /v1/fungible-tokens/{tokenName}/withdrawals`

Parameters

- `tokenName`: The name of the fungible token (e.g. "flow")

---

#### Get details of a token withdrawal

:warning: _Not yet implemented_

`GET /v1/fungible-tokens/{tokenName}/withdrawals/{transactionId}`

Parameters

- `tokenName`: The name of the fungible token (e.g. "flow")
- `transactionId`: The Flow transaction ID for the withdrawal

---

#### Create a token withdrawal

`POST /v1/fungible-tokens/{tokenName}/withdrawals`

Parameters

- `tokenName`: The name of the fungible token (e.g. "flow")

Body (JSON)

- `amount`: The number of tokens to transfer (e.g. "123.456")
  - Must be a fixed-point number with a maximum of 8 decimal places
- `recipient`: The Flow address of the recipient (e.g. "0xf8d6e0586b0a20c7")

Example

```sh
curl --request GET \
  --url http://localhost:3000/v1/fungible-tokens \
  --header 'Content-Type: application/json' \
  --data '{ "recipient": "0xf8d6e0586b0a20c7", "amount": "123.456" }'
```

```json
{
  "transactionId": "18647b584a03345f3b2d2c4d9ab2c4179ae1b124a7f62ef9f33910e5ca8b353c",
  "recipient": "0xf8d6e0586b0a20c7",
  "amount": "123.456"
}
```

---

### Non-Fungible Tokens

:warning: _Not yet implemented_

#### List all tokens

:warning: _Not yet implemented_

`GET /v1/non-fungible-tokens`

Example

```sh
curl --request GET \
  --url http://localhost:3000/v1/non-fungible-tokens
```

---

#### Get details of a token

:warning: _Not yet implemented_

`GET /v1/non-fungible-tokens/{tokenName}`

Parameters

- `tokenName`: The name of the non-fungible token (e.g. "nba-top-shot-moment")

---

#### List all withdrawals of a token type

:warning: _Not yet implemented_

`GET /v1/non-fungible-tokens/{tokenName}/withdrawals`

Parameters

- `tokenName`: The name of the non-fungible token (e.g. "nba-top-shot-moment")

---

#### Get details of a token withdrawal

:warning: _Not yet implemented_

`GET /v1/non-fungible-tokens/{tokenName}/withdrawals/{transactionId}`

Parameters

- `tokenName`: The name of the non-fungible token (e.g. "nba-top-shot-moment")
- `transactionId`: The Flow transaction ID for the withdrawal

---

#### Create a token withdrawal

:warning: _Not yet implemented_

`POST /v1/non-fungible-tokens/{tokenName}/withdrawals`

Parameters

- `tokenName`: The name of the non-fungible token (e.g. "nba-top-shot-moment")

Body (JSON)

- `recipient`: The Flow address of the recipient (e.g. "0xf8d6e0586b0a20c7")

