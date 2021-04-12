const {transferFLOW, transferFUSD} = require("./transfer")

const tokenFLOW = "flow"
const tokenFUSD = "fusd"

const tokens = [tokenFLOW, tokenFUSD]

const transferFuncs = {
  [tokenFLOW]: transferFLOW,
  [tokenFUSD]: transferFUSD,
}

function isValidToken(token) {
  return token in transferFuncs
}

function getTokenTransferFunc(token) {
  return transferFuncs[token]
}

module.exports = {
  tokens,
  isValidToken,
  getTokenTransferFunc,
}
