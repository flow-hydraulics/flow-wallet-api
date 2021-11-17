package chain_events

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"gorm.io/gorm"
)

type GetEventTypes func() ([]string, error)

type Listener struct {
	ticker         *time.Ticker
	done           chan bool
	logger         *log.Logger
	fc             *client.Client
	db             Store
	getTypes       GetEventTypes
	maxBlocks      uint64
	interval       time.Duration
	startingHeight uint64
}

type ListenerStatus struct {
	gorm.Model
	LatestHeight uint64
}

func (ListenerStatus) TableName() string {
	return "chain_events_status"
}

func NewListener(
	logger *log.Logger,
	fc *client.Client,
	db Store,
	getTypes GetEventTypes,
	maxDiff uint64,
	interval time.Duration,
	startingHeight uint64,
) *Listener {
	if logger == nil {
		logger = log.New(os.Stdout, "[EVENT-POLLER] ", log.LstdFlags|log.Lshortfile)
	}
	return &Listener{
		nil, make(chan bool),
		logger, fc, db, getTypes,
		maxDiff, interval, startingHeight,
	}
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

func (l *Listener) handleError(err error) {
	l.logger.Println(err)
	if strings.Contains(err.Error(), "key not found") {
		l.logger.Println(`"key not found" error indicates data is not available at this height, please manually set correct starting height`)
	}
}

func (l *Listener) Start() *Listener {
	if l.ticker != nil {
		// Already started
		return l
	}

	if err := l.initHeight(); err != nil {
		_, ok := err.(*LockError)
		if !ok {
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
				lockErr := l.db.LockedStatus(func(status *ListenerStatus) error {
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

				if lockErr != nil {
					_, ok := lockErr.(*LockError)
					if !ok {
						l.handleError(lockErr)
					}
					// Skip on LockError as it means another listener is already handling this round
				}
			}
		}
	}()

	l.logger.Println("started")

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
	l.logger.Println("stopping...")
	l.ticker.Stop()
	l.done <- true
	l.ticker = nil
}
