package tokens

import (
	"fmt"
	"strings"

	"github.com/eqlabs/flow-wallet-api/accounts"
	"github.com/eqlabs/flow-wallet-api/chain_events"
	"github.com/eqlabs/flow-wallet-api/templates"
	"github.com/onflow/flow-go-sdk"
)

type ChainEventHandler struct {
	AccountService  *accounts.Service
	ChainListener   *chain_events.Listener
	TemplateService *templates.Service
	TokenService    *Service
}

func (h *ChainEventHandler) Handle(event flow.Event) {
	if strings.Contains(event.Type, "Deposit") {
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

	// Check if recipient is in database
	account, err := h.AccountService.Details(accountAddress.String())
	if err != nil {
		return
	}

	if err = h.TokenService.RegisterDeposit(token, event.TransactionID.Hex(), amountOrNftID.String(), account.Address); err != nil {
		fmt.Printf("error while registering a deposit: %s\n", err)
		return
	}
}
