import NonFungibleToken from <BaseNFTAddress>
import <NFTName> from <NFTAddress>

// This transaction configures an account to hold Kitty Items.

transaction {
    prepare(signer: AuthAccount) {
        // if the account doesn't already have a collection
        if signer.borrow<&<NFTName>.Collection>(from: <NFTName>.CollectionStoragePath) == nil {

            // create a new empty collection
            let collection <- <NFTName>.createEmptyCollection()

            // save it to the account
            signer.save(<-collection, to: <NFTName>.CollectionStoragePath)

            // create a public capability for the collection
            signer.link<&<NFTName>.Collection{NonFungibleToken.CollectionPublic, <NFTName>.<NFTName>CollectionPublic}>(<NFTName>.CollectionPublicPath, target: <NFTName>.CollectionStoragePath)
        }
    }
}
