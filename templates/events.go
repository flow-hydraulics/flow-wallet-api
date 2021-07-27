package templates

import (
	"fmt"
	"strings"
)

const (
	EventTokensDeposited = "TokensDeposited" // FungibleToken
	EventDeposit         = "Deposit"         // NonFungibleToken
)

func EventType(address, tokenName, eventName string) string {
	return fmt.Sprintf("A.%s.%s.%s", address, tokenName, eventName)
}

func DepositNameFromTokenType(tokenType string) string {
	switch tokenType {
	default:
		return ""
	case "FT":
		return EventTokensDeposited
	case "NFT":
		return EventDeposit
	}
}

func DepositEventTypeFromToken(token BasicToken) string {
	address := strings.TrimPrefix(token.Address, "0x")
	eventName := DepositNameFromTokenType(token.Type.String())
	return EventType(address, token.Name, eventName)
}
