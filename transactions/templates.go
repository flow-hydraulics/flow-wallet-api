package transactions

const CreateAccountTemplate = `
transaction(publicKeys: [String]) {
	prepare(signer: AuthAccount) {
		let acct = AuthAccount(payer: signer)

		for key in publicKeys {
			acct.addPublicKey(key.decodeHex())
		}
	}
}
`
