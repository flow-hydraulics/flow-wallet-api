const fcl = require("@onflow/fcl");

async function sendTransaction({
  transaction,
  args,
  proposer,
  authorizations,
  payer,
}) {
  const response = await fcl.send([
    fcl.transaction(transaction),
    fcl.args(args),
    fcl.proposer(proposer),
    fcl.authorizations(authorizations),
    fcl.payer(payer),
    fcl.limit(1000),
  ])

  const { transactionId } = response;

  await fcl.tx(response).onceSealed()

  return transactionId
}

module.exports = sendTransaction;
