import * as express from "express"
import FungibleTokensController from "../../../../controllers/fungibleTokens"
import catchAsync from "../../../../errors/catchAsync"
import createWithdrawalsRouter from "./withdrawals"

function createRouter(
  fungibleTokens: FungibleTokensController
): express.Router {
  const router = express.Router({mergeParams: true})

  const withdrawalsRouter = createWithdrawalsRouter(fungibleTokens)

  router
    .route("/")
    .get(catchAsync((req, res) => fungibleTokens.getTokens(req, res)))

  router
    .route("/:tokenName")
    .get(catchAsync((req, res) => fungibleTokens.getToken(req, res)))

  router.use("/:tokenName/withdrawals", withdrawalsRouter)

  return router
}

export default createRouter
