import config from "../config"
import * as fcl from "@onflow/fcl"

import getLeastRecentAdminSignerKey from "../database/getLeastRecentAdminSignerKey"
import getLeastRecentAccountKey from "../database/getLeastRecentAdminSignerKey"
import {AccountAuthorization, AccountAuthorizer, getAuthorization} from "../lib/flow"
import { HashAlgorithm, SignatureAlgorithm } from "../lib/flow/crypto"

interface Account {
  address: string
  sigAlgo: SignatureAlgorithm
  hashAlgo: HashAlgorithm
}

// TODO: add support for user accounts
//
// The current API is hard-coded to only support the admin account
const adminAccount: Account = { 
  address: config.adminAddress,
  sigAlgo: config.adminSigAlgo,
  hashAlgo: config.adminHashAlgo,
}

const allAccounts: Account[] = [adminAccount]

export const queryAccounts = async (): Promise<Account[]> => {
  return allAccounts
}

export const getAccountByAddress = async (address: string): Promise<Account | null> => {
  if (address === adminAccount.address) {
    return adminAccount
  }

  return null
}

const getPrivateKeyByAddress = async (address: string): Promise<string | null> => {
  if (address === adminAccount.address) {
    return config.adminPrivateKey
  }

  return null
}

export const getAccountAuthorization = async (address: string): Promise<AccountAuthorizer> => {

  const account = await getAccountByAddress(address)

  const privateKey = await getPrivateKeyByAddress(address)

  const keyIndex = await getLeastRecentAdminSignerKey()

  return getAuthorization(
    account.address,
    privateKey,
    account.sigAlgo,
    account.hashAlgo,
    keyIndex,
  )
}
