import WalletApiClient from "../../lib/walletApi";
import {adminAddress, baseUrl, fusdTokenName} from "../../lib/config"

const walletApi = new WalletApiClient(baseUrl)

export default async function handler(req, res) {
  if (req.method === "GET") {
    return get(req, res)
  }

  if (req.method === "POST") {
    return post(req, res)
  }
}

async function get(req, res) {
  const accounts = await walletApi.getAccounts()

  const result = await Promise.all(accounts.map(async account => ({
    address: account.address,
    balance: await getFusdBalance(account.address),
    isAdmin: account.address === adminAddress,
  })))

  res.status(200).json(result)
}

const emptyBalance = "0.00000000"

async function post(req, res) {
  const address = await walletApi.createAccount()

  await walletApi.initFungibleToken(address, fusdTokenName)

  const result = {
    address,
    balance: emptyBalance
  }

  res.status(201).json(result)
}

async function getFusdBalance(address) {
  const result = await walletApi.getFungibleToken(address, fusdTokenName)
  return result.balance
}
