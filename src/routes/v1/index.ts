import * as express from "express"

import AccountsController from "src/controllers/accounts"
import FungibleTokensController from "src/controllers/fungibleTokens"
import TransactionsController from "src/controllers/transactions"

import createAccountsRouter from "./accounts"

function createRouter(
  accounts: AccountsController,
  transactions: TransactionsController,
  fungibleTokens: FungibleTokensController
): express.Router {
  const router = express.Router()

  const accountsRouter = createAccountsRouter(
    accounts,
    transactions,
    fungibleTokens
  )

  router.use("/accounts", accountsRouter)

  return router
}

export default createRouter
