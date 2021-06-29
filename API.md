# Wallet API REST HTTP Routes

TODO: this document will contain the REST 
routes provided by the Flow Wallet API.

These routes are also defined in the [OpenAPI specification for this service](openapi.yml).

## Functionality

### Accounts

- [x] Initialize the API with a root admin account
- [x] Create a new Flow account

### Transactions

- [x] Send a transaction signed by an account
- [x] Get the transaction history for an account

### 3. Fungible Tokens

- [x] Send a fungible token withdrawal from an account (FLOW, FUSD)
- [x] Detect fungible token deposits to an account (FLOW, FUSD)
- [x] View the fungible token balance of an account

### 4. Non-Fungible Tokens

- [ ] Set up an account with non-fungible token collections (`NFT.Collection`)
- [ ] Send a non-fungible token withdrawal from an account
- [ ] Detect non-fungible token deposits to an account
- [ ] View the non-fungible tokens owned by an account
