import * as express from "express"

import AccountsController from "src/controllers/accounts"
import FungibleTokensController from "src/controllers/fungibleTokens"
import TransactionsController from "src/controllers/transactions"
import catchAsync from "src/errors/catchAsync"

import createFungibleTokensRouter from "./fungibleTokens"
import createTransactionsRouter from "./transactions"

function createRouter(
  accounts: AccountsController,
  transactions: TransactionsController,
  fungibleTokens: FungibleTokensController
): express.Router {
  const router = express.Router()

  const transactionsRouter = createTransactionsRouter(transactions)
  const fungibleTokensRouter = createFungibleTokensRouter(fungibleTokens)

  router
    .route("/")
    .get(catchAsync((req, res) => accounts.getAccounts(req, res)))
    .post(catchAsync((req, res) => accounts.createAccount(req, res)))

  router
    .route("/:address")
    .get(catchAsync((req, res) => accounts.getAccount(req, res)))

  router.use("/:address/transactions", transactionsRouter)
  router.use("/:address/fungible-tokens", fungibleTokensRouter)

  return router
}

export default createRouter
