import * as express from "express"

import NotFoundError from "src/errors/NotFoundError"
import AccountsService from "src/services/accounts"

export default class AccountsController {
  private accounts: AccountsService

  constructor(accounts: AccountsService) {
    this.accounts = accounts
  }

  async createAccount(
    req: express.Request,
    res: express.Response
  ): Promise<void> {
    const account = await this.accounts.create()

    res.json(account)
  }

  async getAccounts(
    req: express.Request,
    res: express.Response
  ): Promise<void> {
    const accounts = await this.accounts.query()

    res.json(accounts)
  }

  async getAccount(req: express.Request, res: express.Response): Promise<void> {
    const address = req.params.address

    const account = await this.accounts.getByAddress(address)

    if (account === null) {
      throw new NotFoundError()
    }

    res.json(account)
  }
}
