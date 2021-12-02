package tokens

import (
	"github.com/flow-hydraulics/flow-wallet-api/accounts"
	"github.com/flow-hydraulics/flow-wallet-api/flow_helpers"
	"github.com/flow-hydraulics/flow-wallet-api/templates"
	log "github.com/sirupsen/logrus"
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
		log.
			WithFields(log.Fields{"error": err}).
			Warn("Error while adding FlowToken to new account")
	}

	// No need to setup FlowToken

	err = h.TokenService.store.InsertAccountToken(&AccountToken{
		AccountAddress: address,
		TokenAddress:   token.Address,
		TokenName:      token.Name,
		TokenType:      token.Type,
	})
	if err != nil {
		log.
			WithFields(log.Fields{"error": err}).
			Warn("Error while adding FlowToken to new account")
	}
}
