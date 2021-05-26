package templates

import "github.com/onflow/flow-go-sdk"

func init() {
	t := make(templateVariables, 3)

	t["FUNGIBLE_TOKEN_ADDRESS"] = chainAddresses{
		flow.Emulator: "0xee82856bf20e2aa6",
		flow.Testnet:  "0x9a0766d93b6608b7",
		flow.Mainnet:  "0xf233dcee88fe0abe",
	}

	t["FLOW_TOKEN_ADDRESS"] = chainAddresses{
		flow.Emulator: "0x0ae53cb6e3f42a79",
		flow.Testnet:  "0x7e60df042a9c0868",
		flow.Mainnet:  "0x1654653399040a61",
	}

	t["FUSD_ADDRESS"] = chainAddresses{
		flow.Testnet: "0xe223d8a629e49c68",
	}

	replacers = makeChainReplacers(t)
}
