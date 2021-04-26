import * as express from "express"
import AccountsController from "../../../controllers/accounts"
import FungibleTokensController from "../../../controllers/fungibleTokens"
import catchAsync from "../../../errors/catchAsync"
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

  router
    .route("/:address")
    .get(catchAsync((req, res) => accounts.getAccount(req, res)))

  router.use("/:address/fungible-tokens", fungibleTokensRouter)

  return router
}

export default createRouter
