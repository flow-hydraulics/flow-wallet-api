package templates

import "github.com/onflow/flow-go-sdk"

func init() {
	t := make(templateVariables, 4)

	t[`"./FungibleToken.cdc"`] = chainAddresses{
		flow.Emulator: "0xee82856bf20e2aa6",
		flow.Testnet:  "0x9a0766d93b6608b7",
		flow.Mainnet:  "0xf233dcee88fe0abe",
	}

	t["FUNGIBLE_TOKEN_ADDRESS"] = chainAddresses{
		flow.Emulator: "0xee82856bf20e2aa6",
		flow.Testnet:  "0x9a0766d93b6608b7",
		flow.Mainnet:  "0xf233dcee88fe0abe",
	}

	replacers = makeChainReplacers(t)
}
