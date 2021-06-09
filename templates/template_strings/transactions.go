package template_strings

const AddAccountContractWithAdmin = `
transaction(name: String, code: String) {
	prepare(signer: AuthAccount) {
		signer.contracts.add(name: name, code: code.decodeHex(), adminAccount: signer)
	}
}
`

const GenericFungibleTransfer = `
import FungibleToken from FUNGIBLE_TOKEN_ADDRESS
import TokenName from TOKEN_NAME_ADDRESS

transaction(amount: UFix64, recipient: Address) {
  let sentVault: @FungibleToken.Vault

  prepare(signer: AuthAccount) {
    let vaultRef = signer
      .borrow<&TokenName.Vault>(from: /storage/tokenNameVault)
      ?? panic("failed to borrow reference to sender vault")

    self.sentVault <- vaultRef.withdraw(amount: amount)
  }

  execute {
    let receiverRef = getAccount(recipient)
      .getCapability(/public/tokenNameReceiver)
      .borrow<&{FungibleToken.Receiver}>()
      ?? panic("failed to borrow reference to recipient vault")

    receiverRef.deposit(from: <-self.sentVault)
  }
}
`

const GenericFungibleSetup = `
import FungibleToken from FUNGIBLE_TOKEN_ADDRESS
import TokenName from TOKEN_NAME_ADDRESS

transaction {
  prepare(signer: AuthAccount) {

    let existingVault = signer.borrow<&TokenName.Vault>(from: /storage/tokenNameVault)

    if (existingVault != nil) {
        panic("vault exists")
    }

    signer.save(<-TokenName.createEmptyVault(), to: /storage/tokenNameVault)

    signer.link<&TokenName.Vault{FungibleToken.Receiver}>(
      /public/tokenNameReceiver,
      target: /storage/tokenNameVault
    )

    signer.link<&TokenName.Vault{FungibleToken.Balance}>(
      /public/tokenNameBalance,
      target: /storage/tokenNameVault
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

const TransferFlow = `
import FungibleToken from FUNGIBLE_TOKEN_ADDRESS
import FlowToken from FLOW_TOKEN_ADDRESS

transaction(amount: UFix64, recipient: Address) {
  let sentVault: @FungibleToken.Vault

  prepare(signer: AuthAccount) {
    let vaultRef = signer
      .borrow<&FlowToken.Vault>(from: /storage/flowTokenVault)
      ?? panic("failed to borrow reference to sender vault")

    self.sentVault <- vaultRef.withdraw(amount: amount)
  }

  execute {
    let receiverRef = getAccount(recipient)
      .getCapability(/public/flowTokenReceiver)
      .borrow<&{FungibleToken.Receiver}>()
      ?? panic("failed to borrow reference to recipient vault")

    receiverRef.deposit(from: <-self.sentVault)
  }
}
`

const TransferFUSD = `
import FungibleToken from FUNGIBLE_TOKEN_ADDRESS
import FUSD from FUSD_ADDRESS

transaction(amount: UFix64, recipient: Address) {
  let sentVault: @FungibleToken.Vault

  prepare(signer: AuthAccount) {
    let vaultRef = signer
      .borrow<&FUSD.Vault>(from: /storage/fusdVault)
      ?? panic("failed to borrow reference to sender vault")

    self.sentVault <- vaultRef.withdraw(amount: amount)
  }

  execute {
    let receiverRef = getAccount(recipient)
      .getCapability(/public/fusdReceiver)
      .borrow<&{FungibleToken.Receiver}>()
      ?? panic("failed to borrow reference to recipient vault")

    receiverRef.deposit(from: <-self.sentVault)
  }
}
`

const SetupFUSD = `
import FungibleToken from FUNGIBLE_TOKEN_ADDRESS
import FUSD from FUSD_ADDRESS

transaction {
  prepare(signer: AuthAccount) {

    let existingVault = signer.borrow<&FUSD.Vault>(from: /storage/fusdVault)

    if (existingVault != nil) {
        return
    }

    signer.save(<-FUSD.createEmptyVault(), to: /storage/fusdVault)

    signer.link<&FUSD.Vault{FungibleToken.Receiver}>(
      /public/fusdReceiver,
      target: /storage/fusdVault
    )

    signer.link<&FUSD.Vault{FungibleToken.Balance}>(
      /public/fusdBalance,
      target: /storage/fusdVault
    )
  }
}
`
