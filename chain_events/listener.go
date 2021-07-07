package chain_events

import (
	"context"
	"fmt"
	"time"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"gorm.io/gorm"
)

type GetEventTypes func() []string

type Listener struct {
	ticker   *time.Ticker
	done     chan bool
	fc       *client.Client
	db       Store
	maxDiff  uint64
	interval time.Duration
	getTypes GetEventTypes
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
	maxDiff uint64,
	interval time.Duration,
	getTypes GetEventTypes,
) *Listener {
	return &Listener{nil, make(chan bool), fc, db, maxDiff, interval, getTypes}
}

func (l *Listener) run(ctx context.Context, start, end uint64) error {
	ee := make([]flow.Event, 0)

	types := l.getTypes()

	for _, t := range types {
		r, err := l.fc.GetEventsForHeightRange(ctx, client.EventRangeQuery{
			Type:        t,
			StartHeight: start,
			EndHeight:   end,
		})
		if err != nil {
			return fmt.Errorf("error while fetching events: %w", err)
		}
		for _, b := range r {
			ee = append(ee, b.Events...)
		}
	}

	for _, event := range ee {
		Event.Trigger(event)
	}

	return nil
}

func (l *Listener) handleError(err error) {
	fmt.Println(err)
}

func (l *Listener) Start() *Listener {
	if l.ticker != nil {
		return l
	}

	l.ticker = time.NewTicker(l.interval)

	go func() {
		ctx := context.Background()
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		status, err := l.db.GetListenerStatus()
		if err != nil {
			panic(err)
		}

		for {
			select {
			case <-l.done:
				return
			case <-l.ticker.C:
				cur, err := l.fc.GetLatestBlock(ctx, true)
				if err != nil {
					l.handleError(err)
					continue
				}
				curHeight := cur.Height
				if curHeight > status.LatestHeight {
					start := status.LatestHeight + 1       // latestHeight has already been checked, add 1
					end := min(curHeight, start+l.maxDiff) // Limit maximum end
					if err := l.run(ctx, start, end); err != nil {
						l.handleError(err)
						continue
					}
					status.LatestHeight = end
					err := l.db.UpdateListenerStatus(status)
					if err != nil {
						l.handleError(err)
						continue
					}
				}
			}
		}
	}()

	return l
}

func (l *Listener) Stop() {
	l.ticker.Stop()
	l.done <- true
	l.ticker = nil
}

func min(x, y uint64) uint64 {
	if x > y {
		return y
	}
	return x
}
