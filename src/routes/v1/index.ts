import * as express from "express"

import AccountsController from "src/controllers/accounts"
import FungibleTokensController from "src/controllers/fungibleTokens"

import createAccountsRouter from "./accounts"

function createRouter(
  accounts: AccountsController,
  fungibleTokens: FungibleTokensController
): express.Router {
  const router = express.Router()

  const accountsRouter = createAccountsRouter(accounts, fungibleTokens)

  router.use("/accounts", accountsRouter)

  return router
}

export default createRouter
