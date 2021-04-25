import config from "../config"
import * as fcl from "@onflow/fcl"

import getLeastRecentAdminSignerKey from "../database/getLeastRecentAdminSignerKey"
import {AccountAuthorizer, getAuthorization} from "../lib/flow"
import * as Crypto from "../lib/flow/crypto"

interface Account {
  address: string
}

// TODO: add support for user accounts
//
// The current API is hard-coded to only support the admin account
const adminAccount: Account = {
  address: config.adminAddress,
}

const allAccounts: Account[] = [adminAccount]

export const queryAccounts = async (): Promise<Account[]> => {
  return allAccounts
}

export const getAccountByAddress = async (
  address: string
): Promise<Account | null> => {
  if (address === adminAccount.address) {
    return adminAccount
  }

  return null
}

const getSignerByAddress = async (
  address: string
): Promise<Crypto.Signer | null> => {
  // TODO: add support for user accounts
  if (address !== adminAccount.address) {
    return null
  }

  const privateKey = Crypto.InMemoryPrivateKey.fromHex(
    config.adminPrivateKey,
    config.adminSigAlgo
  )

  return new Crypto.InMemorySigner(privateKey, config.adminHashAlgo)
}

export const getAccountAuthorization = async (
  address: string
): Promise<AccountAuthorizer> => {
  const account = await getAccountByAddress(address)

  const signer = await getSignerByAddress(address)

  const keyIndex = await getLeastRecentAdminSignerKey()

  return getAuthorization(account.address, keyIndex, signer)
}
