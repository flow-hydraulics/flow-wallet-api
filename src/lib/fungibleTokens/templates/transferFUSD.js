const dedent = require("dedent-js")

function template(contracts) {
  return dedent`
  import FungibleToken from ${contracts.FungibleToken}
  import FUSD from ${contracts.FUSD}

  transaction(recipient: Address, amount: UFix64) {

    let transferVault: @FungibleToken.Vault

    prepare(signer: AuthAccount) {
      let vaultRef = signer
        .borrow<&FUSD.Vault>(from: /storage/fusdVault)!

      self.transferVault <- vaultRef.withdraw(amount: amount)
    }

    execute {
      let receiverRef = getAccount(recipient)
        .getCapability(/public/fusdReceiver)
        .borrow<&{FungibleToken.Receiver}>()!

      receiverRef.deposit(from: <-self.transferVault)
    }
  }
  `
}

module.exports = template
