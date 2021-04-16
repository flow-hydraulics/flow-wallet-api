import * as httpStatus from "http-status"
import config from "../config"
import {
  tokens,
  isValidToken,
  getTokenTransferFunc,
} from "../lib/fungibleTokens"
import getSigner from "../lib/flow/getSigner"
import getLeastRecentAccountKey from "../database/getLeastRecentAccountKey"
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
  const tokenName = req.params.tokenName

  if (!isValidToken(tokenName)) {
    throw new InvalidFungibleTokenError(tokenName)
  }

  // TODO: validate recipient and amount
  const {recipient, amount} = req.body

  const adminKeyIndex = await getLeastRecentAccountKey()

  const signer = getSigner(
    config.adminAddress,
    config.adminPrivateKey,
    config.adminSigAlgo,
    config.adminHashAlgo,
    adminKeyIndex
  )

  const transfer = getTokenTransferFunc(tokenName)

  try {
    const transactionId = await transfer(
      recipient,
      amount,
      signer,
      config.contracts
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
})
