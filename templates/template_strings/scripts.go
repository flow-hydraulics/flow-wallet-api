package template_strings

const GetFlowBalance = `
import FungibleToken from FUNGIBLE_TOKEN_ADDRESS
import FlowToken from FLOW_TOKEN_ADDRESS

pub fun main(account: Address): UFix64 {

    let vaultRef = getAccount(account)
        .getCapability(/public/flowTokenBalance)
        .borrow<&FlowToken.Vault{FungibleToken.Balance}>()
        ?? panic("Could not borrow Balance reference to the Vault")

    return vaultRef.balance
}
`

const GenericFungibleBalance = `
import FungibleToken from FUNGIBLE_TOKEN_ADDRESS
import TokenName from TOKEN_NAME_ADDRESS

pub fun main(account: Address): UFix64 {

    let vaultRef = getAccount(account)
        .getCapability(/public/tokenNameBalance)
        .borrow<&TokenName.Vault{FungibleToken.Balance}>()
        ?? panic("failed to borrow reference to vault")

    return vaultRef.balance
}
`
