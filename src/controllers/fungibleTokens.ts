import * as httpStatus from "http-status"

import * as express from "express"

import {isValidToken} from "src/lib/fungibleTokens"
import FungibleTokensService from "src/services/fungibleTokens"
import ApiError from "src/errors/ApiError"
import InvalidFungibleTokenError from "src/errors/InvalidFungibleTokenError"

export default class FungibleTokensController {
  private fungibleTokens: FungibleTokensService

  constructor(fungibleTokens: FungibleTokensService) {
    this.fungibleTokens = fungibleTokens
  }

  async getTokens(req: express.Request, res: express.Response): Promise<void> {
    const tokens = await this.fungibleTokens.query()

    res.json(tokens)
  }

  async getToken(req: express.Request, res: express.Response): Promise<void> {
    const tokenName = req.params.tokenName

    if (!isValidToken(tokenName)) {
      throw new InvalidFungibleTokenError(tokenName)
    }

    const token = await this.fungibleTokens.getByName(tokenName)

    res.json(token)
  }

  // TODO: implement withdrawal getters
  async getWithdrawals(
    req: express.Request,
    res: express.Response
  ): Promise<void> {
    res.send("TODO: implement me")
  }

  // TODO: implement withdrawal getters
  async getWithdrawal(
    req: express.Request,
    res: express.Response
  ): Promise<void> {
    res.send("TODO: implement me")
  }

  async createWithdrawal(
    req: express.Request,
    res: express.Response
  ): Promise<void> {
    const sender = req.params.address
    const tokenName = req.params.tokenName

    if (!isValidToken(tokenName)) {
      throw new InvalidFungibleTokenError(tokenName)
    }

    // TODO: validate recipient and amount
    const {recipient, amount} = req.body

    try {
      const transactionId = await this.fungibleTokens.createWithdrawal(
        sender,
        recipient,
        tokenName,
        amount
      )

      const response = {
        transactionId,
        recipient,
        amount,
      }

      res.json(response)
    } catch (e) {
      throw new ApiError(
        httpStatus.INTERNAL_SERVER_ERROR,
        "failed to complete withdrawal"
      )
    }
  }
}
