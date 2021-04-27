import * as fcl from "@onflow/fcl"
import * as types from "@onflow/types"

import {AccountAuthorizer} from "./index"

export type Argument = {
  type: string
  value: string
}

export type Transaction = {
  code: string
  args: Argument[]
  proposer: AccountAuthorizer
  authorizations: AccountAuthorizer[]
  payer: AccountAuthorizer
}

export interface Event {
  data: any // eslint-disable-line
}

export type TransactionResult = {
  id: string
  error: string
  events: Event[]
}

export default async function sendTransaction({
  code,
  args,
  proposer,
  authorizations,
  payer,
}: Transaction): Promise<TransactionResult> {
  const response = await fcl.send([
    fcl.transaction(code),
    fcl.args(args.map(arg => fcl.arg(arg.value, types[arg.type]))),
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
