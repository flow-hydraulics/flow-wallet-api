import * as express from "express"
import FungibleTokensController from "../../../../controllers/fungibleTokens"
import catchAsync from "../../../../errors/catchAsync"

function createRouter(
  fungibleTokens: FungibleTokensController
): express.Router {
  const router = express.Router({mergeParams: true})

  router
    .route("/")
    .get(catchAsync((req, res) => fungibleTokens.getWithdrawals(req, res)))
    .post(catchAsync((req, res) => fungibleTokens.createWithdrawal(req, res)))

  router
    .route("/:transactionId")
    .get(catchAsync((req, res) => fungibleTokens.getWithdrawal(req, res)))

  return router
}

export default createRouter
