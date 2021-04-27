import * as fcl from "@onflow/fcl"

import {AccountAuthorizer} from "../flow"
import sendTransaction, {TransactionResult} from "../flow/sendTransaction"

import transferFLOWTemplate from "./templates/transferFLOW"
import transferFUSDTemplate from "./templates/transferFUSD"

const tokenFLOW = "flow"
const tokenFUSD = "fusd"

export const tokens = [tokenFLOW, tokenFUSD]

type TransactionTemplate = (contracts: {[key: string]: string}) => string

export function isValidToken(tokenName: string): boolean {
  return tokenName == tokenFLOW || tokenName == tokenFUSD
}

function getTransferTemplate(tokenName: string): TransactionTemplate {
  switch (tokenName) {
    case tokenFLOW:
      return transferFLOWTemplate
    case tokenFUSD:
      return transferFUSDTemplate
  }
}

export async function transfer(
  tokenName: string,
  recipient: string,
  amount: string,
  authorization: AccountAuthorizer,
  contracts: {[key: string]: string}
): Promise<TransactionResult> {
  const template = getTransferTemplate(tokenName)

  return await sendTransaction({
    code: template(contracts),
    args: [
      {
        type: "Address",
        value: fcl.withPrefix(recipient),
      },
      {
        type: "UFix64",
        value: amount,
      },
    ],
    authorizations: [authorization],
    payer: authorization,
    proposer: authorization,
  })
}
