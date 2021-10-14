package tokens

import (
	"fmt"
	"strings"

	"github.com/flow-hydraulics/flow-wallet-api/accounts"
	"github.com/flow-hydraulics/flow-wallet-api/chain_events"
	"github.com/flow-hydraulics/flow-wallet-api/flow_helpers"
	"github.com/flow-hydraulics/flow-wallet-api/templates"
	"github.com/onflow/flow-go-sdk"
)

type ChainEventHandler struct {
	AccountService  *accounts.Service
	ChainListener   *chain_events.Listener
	TemplateService *templates.Service
	TokenService    *Service
}

func (h *ChainEventHandler) Handle(event flow.Event) {
	isDeposit := strings.Contains(event.Type, "Deposit")
	if isDeposit {
		h.handleDeposit(event)
	}
}

func (h *ChainEventHandler) handleDeposit(event flow.Event) {
	// We don't have to care about tokens that are not in the database
	// as we could not even listen to events for them
	token, err := h.TemplateService.TokenFromEvent(event)
	if err != nil {
		return
	}

	amountOrNftID := event.Value.Fields[0]
	accountAddress := event.Value.Fields[1]

	// Get the target account from database
	account, err := h.AccountService.Details(flow_helpers.HexString(accountAddress.String()))
	if err != nil {
		return
	}

	if err = h.TokenService.RegisterDeposit(token, event.TransactionID, account, amountOrNftID.String()); err != nil {
		fmt.Printf("Error while registering a deposit: %s\n", err)
		return
	}
}
