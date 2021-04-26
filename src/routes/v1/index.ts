import * as express from "express"
import AccountsController from "../../controllers/accounts"
import FungibleTokensController from "../../controllers/fungibleTokens"
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
