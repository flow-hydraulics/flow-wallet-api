import * as httpStatus from "http-status"

import * as express from "express"

import TransactionsService from "src/services/transactions"
import ApiError from "src/errors/ApiError"

export default class TransactionsController {
  private transactions: TransactionsService

  constructor(transactions: TransactionsService) {
    this.transactions = transactions
  }

  // TODO: implement transaction getters
  async getTransactions(
    req: express.Request,
    res: express.Response
  ): Promise<void> {
    res.send("TODO: implement me")
  }

  // TODO: implement transaction getters
  async getTransaction(
    req: express.Request,
    res: express.Response
  ): Promise<void> {
    res.send("TODO: implement me")
  }

  async createTransaction(
    req: express.Request,
    res: express.Response
  ): Promise<void> {
    const sender = req.params.address

    // TODO: validate code and arguments
    const {code, arguments: args} = req.body

    try {
      const transactionId = await this.transactions.createTransaction(
        sender,
        code,
        args
      )

      const response = {
        transactionId,
      }

      res.json(response)
    } catch (e) {
      throw new ApiError(
        httpStatus.INTERNAL_SERVER_ERROR,
        "failed to complete transaction"
      )
    }
  }
}
