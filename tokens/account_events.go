package tokens

import (
	"fmt"

	"github.com/eqlabs/flow-wallet-api/accounts"
	"github.com/eqlabs/flow-wallet-api/flow_helpers"
	"github.com/eqlabs/flow-wallet-api/templates"
)

type AccountAddedHandler struct {
	TemplateService *templates.Service
	TokenService    *Service
}

func (h *AccountAddedHandler) Handle(payload accounts.AccountAddedPayload) {
	address := flow_helpers.FormatAddress(payload.Address)
	h.addFlowToken(address)
}

func (h *AccountAddedHandler) addFlowToken(address string) {
	token, err := h.TemplateService.GetTokenByName("FlowToken")
	if err != nil {
		fmt.Printf("Error while adding FlowToken to new account: %s\n", err)
	}

	// No need to setup FlowToken

	err = h.TokenService.store.InsertAccountToken(&AccountToken{
		AccountAddress: address,
		TokenAddress:   token.Address,
		TokenName:      token.Name,
		TokenType:      token.Type,
	})
	if err != nil {
		fmt.Printf("Error while adding FlowToken to new account: %s\n", err)
	}
}
