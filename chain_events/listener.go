package chain_events

import (
	"context"

	"strings"
	"time"

	"github.com/flow-hydraulics/flow-wallet-api/system"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type GetEventTypes func() ([]string, error)

type Listener struct {
	ticker         *time.Ticker
	done           chan bool
	fc             *client.Client
	db             Store
	getTypes       GetEventTypes
	maxBlocks      uint64
	interval       time.Duration
	startingHeight uint64

	systemService *system.Service
}

type ListenerStatus struct {
	gorm.Model
	LatestHeight uint64
}

func (ListenerStatus) TableName() string {
	return "chain_events_status"
}

func NewListener(
	fc *client.Client,
	db Store,
	getTypes GetEventTypes,
	maxDiff uint64,
	interval time.Duration,
	startingHeight uint64,
	opts ...ListenerOption,
) *Listener {

	listener := &Listener{
		ticker:         nil,
		done:           make(chan bool),
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

	return listener
}

func (l *Listener) run(ctx context.Context, start, end uint64) error {
	events := make([]flow.Event, 0)

	eventTypes, err := l.getTypes()
	if err != nil {
		return err
	}

	for _, t := range eventTypes {
		r, err := l.fc.GetEventsForHeightRange(ctx, client.EventRangeQuery{
			Type:        t,
			StartHeight: start,
			EndHeight:   end,
		})
		if err != nil {
			return err
		}
		for _, b := range r {
			events = append(events, b.Events...)
		}
	}

	for _, event := range events {
		Event.Trigger(event)
	}

	return nil
}

func (l *Listener) Start() *Listener {
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

		for {
			select {
			case <-l.done:
				return
			case <-l.ticker.C:
				// Check for maintenance mode
				if l.waitMaintenance() {
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
							if strings.Contains(err.Error(), "database is locked") {
								// Sqlite throws this error from time to time when accessing it from
								// multiple threads; listener is run in a separate thread.
								return nil
							}
							return err
						}
						status.LatestHeight = end
					}

					return nil
				})

				if err != nil {
					if _, isLockError := err.(*LockError); isLockError {
						// Skip LockError as it means another listener is already handling this round
						continue
					}

					log.
						WithFields(log.Fields{"error": err}).
						Warn("Error while handling Flow events")

					if strings.Contains(err.Error(), "key not found") {
						log.Warn(`"key not found" error indicates data is not available at this height, please manually set correct starting height`)
					}
				}
			}
		}
	}()

	log.Info("Started")

	return l
}

func (l *Listener) initHeight() error {
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

func (l *Listener) Stop() {
	log.Info("Stopping")
	if l.ticker != nil {
		l.ticker.Stop()
	}
	if l.done != nil {
		l.done <- true
	}
	l.ticker = nil
}

func (l *Listener) waitMaintenance() bool {
	return l.systemService != nil && l.systemService.IsMaintenanceMode()
}
