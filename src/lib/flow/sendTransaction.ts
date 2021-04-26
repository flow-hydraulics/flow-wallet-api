import * as fcl from "@onflow/fcl"

import {AccountAuthorizer} from "./index"

interface Argument {
  value: string
  xform: any // eslint-disable-line
}

type Transaction = {
  transaction: string
  args: Argument[]
  proposer: AccountAuthorizer
  authorizations: AccountAuthorizer[]
  payer: AccountAuthorizer
}

interface Event {
  data: any // eslint-disable-line
}

type TransactionResult = {
  id: string
  error: string
  events: Event[]
}

export default async function sendTransaction({
  transaction,
  args,
  proposer,
  authorizations,
  payer,
}: Transaction): Promise<TransactionResult> {
  const response = await fcl.send([
    fcl.transaction(transaction),
    fcl.args(args),
    fcl.proposer(proposer),
    fcl.authorizations(authorizations),
    fcl.payer(payer),
    fcl.limit(1000),
  ])

  const {transactionId} = response
  const {error, events} = await fcl.tx(response).onceSealed()

  return {
    id: transactionId,
    error: error,
    events: events,
  }
}
