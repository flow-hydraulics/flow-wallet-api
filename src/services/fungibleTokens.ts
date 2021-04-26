import {PrismaClient} from ".prisma/client"
import config from "../config"
import {getTokenTransferFunc, tokens} from "../lib/fungibleTokens"
import Service from "./service"
import AccountsService from "./accounts"

type FungibleToken = {
  name: string
}

const makeToken = tokenName => ({name: tokenName})
const allTokens = tokens.map(tokenName => makeToken(tokenName))

export default class FungibleTokensService extends Service {
  private accounts: AccountsService

  constructor(prisma: PrismaClient, accounts: AccountsService) {
    super(prisma)
    this.accounts = accounts
  }

  async query(): Promise<FungibleToken[]> {
    return allTokens
  }

  async getByName(tokenName: string): Promise<FungibleToken> {
    return makeToken(tokenName)
  }

  async createWithdrawal(
    sender: string,
    recipient: string,
    tokenName: string,
    amount: string
  ): Promise<string> {
    const authorization = await this.accounts.getAuthorization(sender)

    const transfer = getTokenTransferFunc(tokenName)

    const transactionId = await transfer(
      recipient,
      amount,
      authorization,
      config.contracts
    )

    return transactionId
  }
}
