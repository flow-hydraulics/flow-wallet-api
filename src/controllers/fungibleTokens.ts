import * as httpStatus from "http-status"
import config from "../config"
import {tokens, isValidToken, getTokenTransferFunc} from "../lib/fungibleTokens"
import {getAccountAuthorization} from "../services/accounts"
import catchAsync from "../errors/catchAsync"
import ApiError from "../errors/ApiError"
import InvalidFungibleTokenError from "../errors/InvalidFungibleTokenError"

const makeToken = tokenName => ({name: tokenName})
const allTokens = tokens.map(tokenName => makeToken(tokenName))

export const getTokens = catchAsync(async (req, res) => res.json(allTokens))

export const getToken = catchAsync(async (req, res) => {
  const tokenName = req.params.tokenName

  if (!isValidToken(tokenName)) {
    throw new InvalidFungibleTokenError(tokenName)
  }

  const token = makeToken(tokenName)

  res.json(token)
})

// TODO: implement withdrawal getters
export const getWithdrawals = catchAsync(async (req, res) => {
  res.send("TODO: implement me")
})

// TODO: implement withdrawal getters
export const getWithdrawal = catchAsync(async (req, res) => {
  res.send("TODO: implement me")
})

export const createWithdrawal = catchAsync(async (req, res) => {
  const address = req.params.address
  const tokenName = req.params.tokenName

  if (!isValidToken(tokenName)) {
    throw new InvalidFungibleTokenError(tokenName)
  }

  // TODO: validate recipient and amount
  const {recipient, amount} = req.body

  const authorization = await getAccountAuthorization(address)

  const transfer = getTokenTransferFunc(tokenName)

  try {
    const transactionId = await transfer(
      recipient,
      amount,
      authorization,
      config.contracts
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
})
