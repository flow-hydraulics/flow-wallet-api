package chain_events

import (
	"github.com/flow-hydraulics/flow-wallet-api/system"
)

type ListenerOption func(*ListenerImpl)

func WithSystemService(svc system.Service) ListenerOption {
	return func(listener *ListenerImpl) {
		listener.systemService = svc
	}
}
