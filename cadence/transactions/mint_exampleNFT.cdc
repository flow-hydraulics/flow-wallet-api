import NonFungibleToken from "../contracts/NonFungibleToken.cdc"
import ExampleNFT from "../contracts/ExampleNFT.cdc"

transaction(recipient: Address) {

    let minter: &ExampleNFT.NFTMinter

    prepare(signer: AuthAccount) {
        self.minter = signer
            .borrow<&ExampleNFT.NFTMinter>(from: ExampleNFT.MinterStoragePath)!
    }

    execute {
        let receiver = getAccount(recipient)
            .getCapability(ExampleNFT.CollectionPublicPath)!
            .borrow<&{NonFungibleToken.CollectionPublic}>()!

        self.minter.mintNFT(recipient: receiver)
    }
}
