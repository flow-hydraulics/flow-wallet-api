import {PrismaClient} from "@prisma/client"

import sendTransaction, {Argument} from "src/lib/flow/sendTransaction"

import Service from "./service"
import AccountsService from "./accounts"

export default class TransactionsService extends Service {
  private accounts: AccountsService

  constructor(prisma: PrismaClient, accounts: AccountsService) {
    super(prisma)
    this.accounts = accounts
  }

  async createTransaction(
    signer: string,
    code: string,
    args: Argument[]
  ): Promise<string> {
    const userAuthorization = await this.accounts.getAuthorization(signer)

    const {id} = await sendTransaction({
      code,
      args,
      proposer: userAuthorization,
      authorizations: [userAuthorization],
      payer: userAuthorization,
    })

    return id
  }
}
