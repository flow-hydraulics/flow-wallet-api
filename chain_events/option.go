package chain_events

import (
	"github.com/flow-hydraulics/flow-wallet-api/system"
)

type ListenerOption func(*Listener)

func WithSystemService(svc *system.Service) ListenerOption {
	return func(listener *Listener) {
		listener.systemService = svc
	}
}
