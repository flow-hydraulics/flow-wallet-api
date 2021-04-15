package deposit

import "github.com/onflow/flow-go-sdk"

// TODO: Deposit external users NFT

// https://flowscan.org/contract/A.1d7e57aa55817448.NonFungibleToken

type NFTDeposition struct {
	chainId            flow.ChainID
	nftContractAddress flow.Address
	nftContractName    string
	nftId              uint64
	extUserId          string
}
