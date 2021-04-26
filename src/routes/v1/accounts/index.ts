import * as express from "express"

import AccountsController from "src/controllers/accounts"
import FungibleTokensController from "src/controllers/fungibleTokens"
import catchAsync from "src/errors/catchAsync"

import createFungibleTokensRouter from "./fungibleTokens"

function createRouter(
  accounts: AccountsController,
  fungibleTokens: FungibleTokensController
): express.Router {
  const router = express.Router()

  const fungibleTokensRouter = createFungibleTokensRouter(fungibleTokens)

  router
    .route("/")
    .get(catchAsync((req, res) => accounts.getAccounts(req, res)))
    .post(catchAsync((req, res) => accounts.createAccount(req, res)))

  router
    .route("/:address")
    .get(catchAsync((req, res) => accounts.getAccount(req, res)))

  router.use("/:address/fungible-tokens", fungibleTokensRouter)

  return router
}

export default createRouter
