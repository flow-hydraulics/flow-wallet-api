package chain_events

import (
	"context"

	"strings"
	"time"

	wallet_errors "github.com/flow-hydraulics/flow-wallet-api/errors"
	"github.com/flow-hydraulics/flow-wallet-api/flow_helpers"
	"github.com/flow-hydraulics/flow-wallet-api/system"
	"github.com/onflow/flow-go-sdk"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type GetEventTypes func() ([]string, error)

type Listener interface {
	Start() Listener
	Stop()
}

type ListenerImpl struct {
	ticker         *time.Ticker
	stopChan       chan struct{}
	fc             flow_helpers.FlowClient
	db             Store
	getTypes       GetEventTypes
	maxBlocks      uint64
	interval       time.Duration
	startingHeight uint64

	systemService system.Service
}

type ListenerStatus struct {
	gorm.Model
	LatestHeight uint64
}

func (ListenerStatus) TableName() string {
	return "chain_events_status"
}

func NewListener(
	fc flow_helpers.FlowClient,
	db Store,
	getTypes GetEventTypes,
	maxDiff uint64,
	interval time.Duration,
	startingHeight uint64,
	opts ...ListenerOption,
) Listener {

	listener := &ListenerImpl{
		ticker:         nil,
		stopChan:       make(chan struct{}),
		fc:             fc,
		db:             db,
		getTypes:       getTypes,
		maxBlocks:      maxDiff,
		interval:       interval,
		startingHeight: startingHeight,
		systemService:  nil,
	}

	// Go through options
	for _, opt := range opts {
		opt(listener)
	}

	log.Debug(listener)

	return listener
}

func (l *ListenerImpl) run(ctx context.Context, start, end uint64) error {
	events := make([]flow.Event, 0)

	eventTypes, err := l.getTypes()
	if err != nil {
		return err
	}

	for _, t := range eventTypes {
		r, err := l.fc.GetEventsForHeightRange(ctx, t, start, end)
		if err != nil {
			return err
		}
		count := 0
		for _, b := range r {
			count += len(b.Events)
			events = append(events, b.Events...)
		}
		log.
			WithFields(log.Fields{
				"type":        t,
				"startHeight": start,
				"endHeight":   end,
				"resultCount": count,
			}).
			Debug("Fetching events")
	}

	for _, event := range events {
		ChainEvent.Trigger(ctx, event)
	}

	return nil
}

func (l *ListenerImpl) Start() Listener {
	if l.ticker != nil {
		// Already started
		return l
	}

	if err := l.initHeight(); err != nil {
		if _, isLockError := err.(*LockError); !isLockError {
			panic(err)
		}
		// Skip LockError as it means another listener is already handling this
	}

	// TODO (latenssi): should use random intervals instead
	l.ticker = time.NewTicker(l.interval)

	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		entry := log.WithFields(log.Fields{
			"package":  "chain_events",
			"function": "Listener.Start.goroutine",
		})

		for {
			select {
			case <-l.stopChan:
				return
			case <-l.ticker.C:
				// Check for maintenance mode
				if halted, err := l.systemHalted(); err != nil {
					entry.
						WithFields(log.Fields{"error": err}).
						Warn("Could not get system settings from DB")
					continue
				} else if halted {
					entry.Debug("System halted")
					continue
				}

				err := l.db.LockedStatus(func(status *ListenerStatus) error {
					latestBlock, err := l.fc.GetLatestBlockHeader(ctx, true)
					if err != nil {
						return err
					}

					if latestBlock.Height > status.LatestHeight {
						start := status.LatestHeight + 1                  // LatestHeight has already been checked, add 1
						end := min(latestBlock.Height, start+l.maxBlocks) // Limit maximum end
						if err := l.run(ctx, start, end); err != nil {
							return err
						}
						status.LatestHeight = end
					}

					return nil
				})

				if err != nil {
					if wallet_errors.IsChainConnectionError(err) {
						// Unable to connect to chain, pause system.
						if l.systemService != nil {
							entry.Warn("Unable to connect to chain, pausing system")
							entry.Warn(err)
							if err := l.systemService.Pause(); err != nil {
								entry.
									WithFields(log.Fields{"error": err}).
									Warn("Unable to pause system")
							}
						} else {
							entry.Warn("Unable to connect to chain")
						}
						continue
					}

					entry.
						WithFields(log.Fields{"error": err}).
						Warn("Error while handling Flow events")

					if strings.Contains(err.Error(), "key not found") {
						entry.Warn(`"key not found" error indicates data is not available at this height, please manually set correct starting height`)
					}
				}
			}
		}
	}()

	log.Debug("Started Flow event listener")

	return l
}

func (l *ListenerImpl) initHeight() error {
	return l.db.LockedStatus(func(status *ListenerStatus) error {
		if l.startingHeight > 0 && status.LatestHeight < l.startingHeight-1 {
			status.LatestHeight = l.startingHeight - 1
		}

		if status.LatestHeight == 0 {
			// If starting fresh, we need to start from the latest block as we can't
			// know what is the root of the current spork.
			// Data on Flow is only accessible for the current spork height.
			latestBlock, err := l.fc.GetLatestBlockHeader(context.Background(), true)
			if err != nil {
				return err
			}
			status.LatestHeight = latestBlock.Height
		}

		return nil
	})
}

func (l *ListenerImpl) Stop() {
	log.Debug("Stopping Flow event listener")

	close(l.stopChan)

	if l.ticker != nil {
		l.ticker.Stop()
	}
}

func (l *ListenerImpl) systemHalted() (bool, error) {
	if l.systemService != nil {
		return l.systemService.IsHalted()
	}
	return false, nil
}
