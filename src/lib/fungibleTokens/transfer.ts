import * as fcl from "@onflow/fcl"
import * as t from "@onflow/types"

import sendTransaction from "../flow/sendTransaction"

import transferFLOWTemplate from "./templates/transferFLOW"
import transferFUSDTemplate from "./templates/transferFUSD"

async function transfer(template, recipient, amount, authorization, contracts) {
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

export async function transferFLOW(
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

export async function transferFUSD(
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
