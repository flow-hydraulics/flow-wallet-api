import {PrismaClient} from "@prisma/client"

import {AccountAuthorizer, getAuthorization} from "src/lib/flow"
import getLeastRecentAdminSignerKey from "src/database/getLeastRecentAdminSignerKey"
import * as models from "src/database/models"
import {
  getAccount,
  getAccountKey,
  insertAccount,
  listAccounts,
} from "src/database/accounts"
import {createAccount} from "src/lib/flow/createAccount"
import {Key, KeyManager} from "src/lib/keys"
import config from "src/config"

import Service from "./service"

const adminAccount: models.Account = {
  address: config.adminAddress,
  createdAt: null,
  updatedAt: null,
}

export default class AccountsService extends Service {
  private adminKey: Key
  private userKeyManager: KeyManager<Key>

  constructor(
    prisma: PrismaClient,
    adminKey: Key,
    userKeyManager: KeyManager<Key>
  ) {
    super(prisma)
    this.adminKey = adminKey
    this.userKeyManager = userKeyManager
  }

  async create(): Promise<models.Account> {
    const adminAuthorization = await this.getAdminAuthorization()

    const userKey = this.userKeyManager.generate()

    const address = await createAccount(
      userKey.getPublicKey(),
      userKey.getSignatureAlgorithm(),
      userKey.getHashAlgorithm(),
      adminAuthorization,
      config.contracts
    )

    const userKeyType = this.userKeyManager.getKeyType()
    const userKeyValue = this.userKeyManager.save(userKey)

    return await insertAccount(this.prisma, address, userKeyType, userKeyValue)
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

  async getAuthorization(address: string): Promise<AccountAuthorizer> {
    if (address === adminAccount.address) {
      return this.getAdminAuthorization()
    }

    const accountKey = await getAccountKey(this.prisma, address)

    const signer = this.userKeyManager.load(accountKey.value).getSigner()

    return getAuthorization(address, accountKey.index, signer)
  }

  private async getAdminAuthorization(): Promise<AccountAuthorizer> {
    const adminKeyIndex = await getLeastRecentAdminSignerKey(this.prisma)

    return getAuthorization(
      adminAccount.address,
      adminKeyIndex,
      this.adminKey.getSigner()
    )
  }
}
