import Crypto

transaction(publicKeys: [Crypto.KeyListEntry], contracts: {String: String}) {
	prepare(signer: AuthAccount) {
		panic("Account initialized with custom script")
	}
}
