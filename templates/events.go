package templates

import (
	"fmt"
	"strings"

	"github.com/eqlabs/flow-wallet-api/events"
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

type TokenEnabledHandler struct {
	ChainListener *events.ChainListener
}

func (h *TokenEnabledHandler) Handle(payload events.TokenEnabledPayload) {
	address := strings.TrimPrefix(payload.TokenAddress, "0x")
	depositName := DepositNameFromTokenType(payload.TokenType)

	// Enable token deposit event
	h.ChainListener.AddType(EventType(
		address,
		payload.TokenName,
		depositName,
	))
}

type TokenDisabledHandler struct {
	ChainListener *events.ChainListener
}

func (h *TokenDisabledHandler) Handle(payload events.TokenDisabledPayload) {
	address := strings.TrimPrefix(payload.TokenAddress, "0x")
	depositName := DepositNameFromTokenType(payload.TokenType)

	// Disable token deposit event
	h.ChainListener.RemoveType(EventType(
		address,
		payload.TokenName,
		depositName,
	))
}
