import NonFungibleToken from "../contracts/NonFungibleToken.cdc"
import ExampleNFT from "../contracts/ExampleNFT.cdc"

pub fun main(account: Address): [UInt64] {
    let receiver = getAccount(account)
        .getCapability(ExampleNFT.CollectionPublicPath)!
        .borrow<&{NonFungibleToken.CollectionPublic}>()!

    return receiver.getIDs()
}
