import * as fcl from "@onflow/fcl"

import {AccountAuthorization} from "./index"

interface Argument {
  value: string
  xform: any // eslint-disable-line
}

interface Transaction {
  transaction: string
  args: Argument[]
  proposer: string
  authorizations: AccountAuthorization[]
  payer: string
}

export default async function sendTransaction({
  transaction,
  args,
  proposer,
  authorizations,
  payer,
}: Transaction): Promise<string> {
  const response = await fcl.send([
    fcl.transaction(transaction),
    fcl.args(args),
    fcl.proposer(proposer),
    fcl.authorizations(authorizations),
    fcl.payer(payer),
    fcl.limit(1000),
  ])

  const {transactionId} = response

  await fcl.tx(response).onceSealed()

  return transactionId
}
