import config from "../config"
import catchAsync from "../errors/catchAsync"
import NotFoundError from "../errors/NotFoundError"
import { getAccountByAddress, queryAccounts } from "../services/accounts"

export const getAccounts = catchAsync(async (req, res) => {
  const accounts = await queryAccounts()

  res.json(accounts)
})

export const getAccount = catchAsync(async (req, res) => {
  const address = req.params.address

  const account = await getAccountByAddress(address)

  if (account === null) {
    throw new NotFoundError()
  }

  res.json(account)
})

// TODO: implement create account function
export const createAccount = catchAsync(async (req, res) => {})
