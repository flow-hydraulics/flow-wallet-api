package template_strings

const AddAccountContractWithAdmin = `
transaction(name: String, code: String) {
	prepare(signer: AuthAccount) {
		signer.contracts.add(name: name, code: code.decodeHex(), adminAccount: signer)
	}
}
`

const GenericFungibleTransfer = `
import FungibleToken from "./FungibleToken.cdc"
import TOKEN_DECLARATION_NAME from TOKEN_ADDRESS

transaction(amount: UFix64, recipient: Address) {
  let sentVault: @FungibleToken.Vault

  prepare(signer: AuthAccount) {
    let vaultRef = signer
      .borrow<&TOKEN_DECLARATION_NAME.Vault>(from: /storage/TOKEN_VAULT)
      ?? panic("failed to borrow reference to sender vault")

    self.sentVault <- vaultRef.withdraw(amount: amount)
  }

  execute {
    let receiverRef = getAccount(recipient)
      .getCapability(/public/TOKEN_RECEIVER)
      .borrow<&{FungibleToken.Receiver}>()
      ?? panic("failed to borrow reference to recipient vault")

    receiverRef.deposit(from: <-self.sentVault)
  }
}
`

const GenericFungibleSetup = `
import FungibleToken from "./FungibleToken.cdc"
import TOKEN_DECLARATION_NAME from TOKEN_ADDRESS

transaction {
  prepare(signer: AuthAccount) {

    let existingVault = signer.borrow<&TOKEN_DECLARATION_NAME.Vault>(from: /storage/TOKEN_VAULT)

    if (existingVault != nil) {
        panic("vault exists")
    }

    signer.save(<-TOKEN_DECLARATION_NAME.createEmptyVault(), to: /storage/TOKEN_VAULT)

    signer.link<&TOKEN_DECLARATION_NAME.Vault{FungibleToken.Receiver}>(
      /public/TOKEN_RECEIVER,
      target: /storage/TOKEN_VAULT
    )

    signer.link<&TOKEN_DECLARATION_NAME.Vault{FungibleToken.Balance}>(
      /public/TOKEN_BALANCE,
      target: /storage/TOKEN_VAULT
    )
  }
}
`

const CreateAccount = `
transaction(publicKeys: [String]) {
	prepare(signer: AuthAccount) {
		let acct = AuthAccount(payer: signer)

		for key in publicKeys {
			acct.addPublicKey(key.decodeHex())
		}
	}
}
`
