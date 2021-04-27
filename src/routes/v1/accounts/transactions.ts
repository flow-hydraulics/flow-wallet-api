import * as express from "express"

import TransactionsController from "src/controllers/transactions"
import catchAsync from "src/errors/catchAsync"

function createRouter(transactions: TransactionsController): express.Router {
  const router = express.Router({mergeParams: true})

  router
    .route("/")
    .get(catchAsync((req, res) => transactions.getTransactions(req, res)))
    .post(catchAsync((req, res) => transactions.createTransaction(req, res)))

  router
    .route("/:transactionId")
    .get(catchAsync((req, res) => transactions.getTransaction(req, res)))

  return router
}

export default createRouter
