import {transferFLOW, transferFUSD} from "./transfer"

const tokenFLOW = "flow"
const tokenFUSD = "fusd"

export const tokens = [tokenFLOW, tokenFUSD]

const transferFuncs = {
  [tokenFLOW]: transferFLOW,
  [tokenFUSD]: transferFUSD,
}

export function isValidToken(token) {
  return token in transferFuncs
}

export function getTokenTransferFunc(token) {
  return transferFuncs[token]
}
