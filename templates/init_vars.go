package templates

import "github.com/onflow/flow-go-sdk"

var KnownAddresses = templateVariables{
	"FungibleToken.cdc": knownAddresses{
		flow.Emulator: "0xee82856bf20e2aa6",
		flow.Testnet:  "0x9a0766d93b6608b7",
		flow.Mainnet:  "0xf233dcee88fe0abe",
	},
	"NonFungibleToken.cdc": knownAddresses{
		flow.Emulator: "0xf8d6e0586b0a20c7",
		flow.Testnet:  "0x631e88ae7f1d7c20",
		flow.Mainnet:  "0x1d7e57aa55817448",
	},
}

func init() {
	knownAddressesReplacers = makeReplacers(KnownAddresses)
}
