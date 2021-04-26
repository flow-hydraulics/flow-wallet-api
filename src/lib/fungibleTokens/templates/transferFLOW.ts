import * as dedent from "dedent-js"

export default function template(contracts: {[key: string]: string}): string {
  return dedent`
  import FungibleToken from ${contracts.FungibleToken}
  import FlowToken from ${contracts.FlowToken}

  transaction(recipient: Address, amount: UFix64) {

    let transferVault: @FungibleToken.Vault

    prepare(signer: AuthAccount) {
      let vaultRef = signer
        .borrow<&FlowToken.Vault>(from: /storage/flowTokenVault)!

      self.transferVault <- vaultRef.withdraw(amount: amount)
    }

    execute {
      let receiverRef = getAccount(recipient)
        .getCapability(/public/flowTokenReceiver)
        .borrow<&{FungibleToken.Receiver}>()!

      receiverRef.deposit(from: <-self.transferVault)
    }
  }
  `
}
