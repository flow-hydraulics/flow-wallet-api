package template_strings

const AddAccountContractWithAdmin = `
transaction(name: String, code: String) {
	prepare(signer: AuthAccount) {
		signer.contracts.add(name: name, code: code.decodeHex(), adminAccount: signer)
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

const GenericFungibleTransfer = `
import FungibleToken from "./FungibleToken.cdc"
import TOKEN_DECLARATION_NAME from TOKEN_ADDRESS

transaction(amount: UFix64, recipient: Address) {
  let sentVault: @FungibleToken.Vault

  prepare(signer: AuthAccount) {
    let vaultRef = signer
      .borrow<&TOKEN_DECLARATION_NAME.Vault>(from: TOKEN_VAULT)
      ?? panic("failed to borrow reference to sender vault")

    self.sentVault <- vaultRef.withdraw(amount: amount)
  }

  execute {
    let receiverRef = getAccount(recipient)
      .getCapability(TOKEN_RECEIVER)
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

    let existingVault = signer.borrow<&TOKEN_DECLARATION_NAME.Vault>(from: TOKEN_VAULT)

    if (existingVault != nil) {
        panic("vault exists")
    }

    signer.save(<-TOKEN_DECLARATION_NAME.createEmptyVault(), to: TOKEN_VAULT)

    signer.link<&TOKEN_DECLARATION_NAME.Vault{FungibleToken.Receiver}>(
      TOKEN_RECEIVER,
      target: TOKEN_VAULT
    )

    signer.link<&TOKEN_DECLARATION_NAME.Vault{FungibleToken.Balance}>(
      TOKEN_BALANCE,
      target: TOKEN_VAULT
    )
  }
}
`

const AddProposalKeyTransaction = `
transaction(adminKeyIndex: Int, numProposalKeys: UInt16) {
  prepare(account: AuthAccount) {
    let key = account.keys.get(keyIndex: adminKeyIndex)!
    var count: UInt16 = 0
    while count < numProposalKeys {
      account.keys.add(
            publicKey: key.publicKey,
            hashAlgorithm: key.hashAlgorithm,
            weight: 0.0
        )
        count = count + 1
    }
  }
}
`

// TODO: sigAlgo & hashAlgo as params, add pre-&post-conditions
const AddAccountKeysTransaction = `
transaction(publicKeys: [String]) {
  prepare(signer: AuthAccount) {
    for pbk in publicKeys {
      let key = PublicKey(
        publicKey: pbk.decodeHex(),
        signatureAlgorithm: SignatureAlgorithm.ECDSA_P256
      )

      signer.keys.add(
        publicKey: key,
        hashAlgorithm: HashAlgorithm.SHA3_256,
        weight: 1000.0
      )
    }
  }
}
`
