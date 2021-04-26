import config from "../config"
import {AccountAuthorizer, getAuthorization} from "../lib/flow"
import * as Crypto from "../lib/crypto"
import Service from "./service"
import * as models from "../database/models"
import getLeastRecentAdminSignerKey from "../database/getLeastRecentAdminSignerKey"
import {
  getAccount,
  getAccountKey,
  insertAccount,
  listAccounts,
} from "../database/accounts"
import {createAccount} from "../lib/flow/createAccount"
import {Key, KeyManager} from "../lib/keys"
import {PrismaClient} from "@prisma/client"

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
      config.contracts,
    )

    const userKeyType = this.userKeyManager.keyType
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
