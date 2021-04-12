const fcl = require("@onflow/fcl");
const t = require("@onflow/types");

const sendTransaction = require("../flow/sendTransaction");

const transferFLOWTemplate = require("./templates/transferFLOW");
const transferFUSDTemplate = require("./templates/transferFUSD");

async function transfer(
  template,
  recipient,
  amount,
  authorization,
  contracts
) {
  return await sendTransaction({
    transaction: template(contracts),
    args: [
      fcl.arg(fcl.withPrefix(recipient), t.Address),
      fcl.arg(amount, t.UFix64),
    ],
    authorizations: [authorization],
    payer: authorization,
    proposer: authorization,
  })
}

async function transferFLOW(
  recipient,
  amount,
  authorization,
  contracts
) {
  return transfer(
    transferFLOWTemplate,
    recipient,
    amount,
    authorization,
    contracts
  )
}

async function transferFUSD(
  recipient,
  amount,
  authorization,
  contracts
) {
  return transfer(
    transferFUSDTemplate,
    recipient,
    amount,
    authorization,
    contracts
  )
}

module.exports = {
  transferFLOW,
  transferFUSD,
};
