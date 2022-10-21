package tokens

import (
	"github.com/flow-hydraulics/flow-wallet-api/accounts"
	"github.com/flow-hydraulics/flow-wallet-api/flow_helpers"
	"github.com/flow-hydraulics/flow-wallet-api/templates"
	log "github.com/sirupsen/logrus"
)

type AccountAddedHandler struct {
	TemplateService templates.Service
	TokenService    Service
}

func (h *AccountAddedHandler) Handle(payload accounts.AccountAddedPayload) {
	address := flow_helpers.FormatAddress(payload.Address)
	h.addToken("FlowToken", address)
	for _, t := range payload.InitializedFungibleTokens {
		h.addToken(t.Name, address)
	}
}

func (h *AccountAddedHandler) addToken(name string, address string) {
	if err := h.TokenService.AddAccountToken(name, address); err != nil {
		log.
			WithFields(log.Fields{"error": err}).
			Warnf("Error while adding %s token to new account", name)
	}
}
