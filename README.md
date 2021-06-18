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
- Transfer NFTs (e.g. FLOW, FUSD)
- Detect NFT deposits

[View full list of supported functionality](#functionality).

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

## Functionality

### 1. Admin

- [x] Single admin account (hot wallet)
- [x] Create user accounts (using admin account)

### 2. Transaction Execution

- [x] Send an arbitrary transaction from the admin account
- [x] Send an arbitrary transaction from a user account

### 3. Fungible Tokens

- [x] Send fungible token withdrawals from admin account (FLOW, FUSD)
- [ ] Detect fungible token deposits to admin account (FLOW, FUSD)
- [x] Send fungible token withdrawals from a user account (FLOW, FUSD)
- [ ] Detect fungible token deposits to a user account (FLOW, FUSD)
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

## Installation

TODO

## API Specification

[View the full Wallet API specification](API.md).
