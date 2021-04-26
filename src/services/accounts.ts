import config from "../config"
import {AccountAuthorizer, getAuthorization} from "../lib/flow"
import * as Crypto from "../lib/crypto"
import Service from "./service"
import * as models from "../database/models"
import getLeastRecentAdminSignerKey from "../database/getLeastRecentAdminSignerKey"
import {getAccount, insertAccount, listAccounts} from "../database/accounts"
import {createAccount} from "../lib/flow/createAccount"

// TODO: add support for user accounts
//
// The current API is hard-coded to only support the admin account
const adminAccount: models.Account = {
  address: config.adminAddress,
  createdAt: null,
  updatedAt: null,
}

const userSigAlgo = Crypto.SignatureAlgorithm.ECDSA_P256
const userHashAlgo = Crypto.HashAlgorithm.SHA3_256

export default class AccountsService extends Service {
  async create(): Promise<models.Account> {
    const adminAuthorization = await this.getAdminAuthorization()

    const userPrivateKey = Crypto.InMemoryPrivateKey.generate(userSigAlgo)
    const userPublicKey = userPrivateKey.getPublicKey()

    const address = await createAccount(
      userPublicKey,
      userSigAlgo,
      userHashAlgo,
      adminAuthorization
    )

    // TODO: save private key
    return await insertAccount(this.prisma, address)
  }

  async query(): Promise<models.Account[]> {
    // TODO: add pagination

    const accounts = await listAccounts(this.prisma)

    return [adminAccount, ...accounts]
  }

  async getByAddress(address: string): Promise<models.Account | null> {
    if (address === adminAccount.address) {
      return adminAccount
    }

    return await getAccount(this.prisma, address)
  }

  async getSignerByAddress(address: string): Promise<Crypto.Signer | null> {
    // TODO: add support for user accounts
    return null
  }

  async getAuthorization(address: string): Promise<AccountAuthorizer> {
    if (address === config.adminAddress) {
      return this.getAdminAuthorization()
    }

    // TODO: add support for user accounts
    return null
  }

  private async getAdminAuthorization(): Promise<AccountAuthorizer> {
    const adminKeyIndex = await getLeastRecentAdminSignerKey(this.prisma)

    const adminPrivateKey = Crypto.InMemoryPrivateKey.fromHex(
      config.adminPrivateKey,
      config.adminSigAlgo
    )

    const adminSigner = new Crypto.InMemorySigner(
      adminPrivateKey,
      config.adminHashAlgo
    )

    return getAuthorization(config.adminAddress, adminKeyIndex, adminSigner)
  }
}
