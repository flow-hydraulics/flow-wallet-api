import config from "../config"
import {AccountAuthorizer, getAuthorization} from "../lib/flow"
import * as Crypto from "../lib/crypto"
import Service from "./service"
import * as models from "../database/models"
import getLeastRecentAdminSignerKey from "../database/getLeastRecentAdminSignerKey"
import {insertAccount} from "../database/accounts"

// TODO: add support for user accounts
//
// The current API is hard-coded to only support the admin account
const adminAccount: models.Account = {
  address: config.adminAddress,
  created_at: null,
  updated_at: null,
}

export default class AccountsService extends Service {
  async create(address: string) {
    await insertAccount(this.prisma, {address: address})
  }

  async query(): Promise<models.Account[]> {
    // TODO: add support for user accounts
    return [adminAccount]
  }

  async getByAddress(address: string): Promise<models.Account | null> {
    // TODO: add support for user accounts
    if (address === adminAccount.address) {
      return adminAccount
    }

    return null
  }

  async getSignerByAddress(address: string): Promise<Crypto.Signer | null> {
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

  async getAuthorization(address: string): Promise<AccountAuthorizer> {
    const account = await this.getByAddress(address)

    const signer = await this.getSignerByAddress(address)

    const keyIndex = await getLeastRecentAdminSignerKey(this.prisma)

    return getAuthorization(account.address, keyIndex, signer)
  }
}
