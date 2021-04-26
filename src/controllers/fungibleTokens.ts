import * as httpStatus from "http-status"
import {isValidToken} from "../lib/fungibleTokens"
import FungibleTokensService from "../services/fungibleTokens"
import ApiError from "../errors/ApiError"
import InvalidFungibleTokenError from "../errors/InvalidFungibleTokenError"

export default class FungibleTokensController {
  private fungibleTokens: FungibleTokensService

  constructor(fungibleTokens: FungibleTokensService) {
    this.fungibleTokens = fungibleTokens
  }

  async getTokens(req, res) {
    const tokens = await this.fungibleTokens.query()

    res.json(tokens)
  }

  async getToken(req, res) {
    const tokenName = req.params.tokenName

    if (!isValidToken(tokenName)) {
      throw new InvalidFungibleTokenError(tokenName)
    }

    const token = await this.fungibleTokens.getByName(tokenName)

    res.json(token)
  }

  // TODO: implement withdrawal getters
  async getWithdrawals(req, res) {
    res.send("TODO: implement me")
  }

  // TODO: implement withdrawal getters
  async getWithdrawal(req, res) {
    res.send("TODO: implement me")
  }

  async createWithdrawal(req, res) {
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
      console.log(e)
      throw new ApiError(
        httpStatus.INTERNAL_SERVER_ERROR,
        "failed to complete withdrawal"
      )
    }
  }
}
