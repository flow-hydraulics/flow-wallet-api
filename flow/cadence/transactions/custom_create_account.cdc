transaction(publicKeys: [String], contracts: {String: String}) {
	prepare(signer: AuthAccount) {
		panic("Account initialized with custom script")
	}
}
