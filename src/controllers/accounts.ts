import NotFoundError from "../errors/NotFoundError"
import AccountsService from "../services/accounts"

export default class AccountsController {
  private accounts: AccountsService

  constructor(accounts: AccountsService) {
    this.accounts = accounts
  }

  async getAccounts(req, res) {
    const accounts = await this.accounts.query()

    res.json(accounts)
  }

  async getAccount(req, res) {
    const address = req.params.address

    const account = await this.accounts.getByAddress(address)

    if (account === null) {
      throw new NotFoundError()
    }

    res.json(account)
  }
}
