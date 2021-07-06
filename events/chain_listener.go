package events

import (
	"context"
	"fmt"
	"time"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"gorm.io/gorm"
)

// TODO: increase once implementation is somewhat complete
const (
	period      = 1 * time.Second
	chanTimeout = period / 2
)

type ChainListener struct {
	ticker  *time.Ticker
	done    chan bool
	types   map[string]bool
	fc      *client.Client
	db      Store
	maxDiff uint64
}

type ListenerStatus struct {
	gorm.Model
	LatestHeight uint64
}

func NewChainListener(fc *client.Client, db Store, maxDiff uint64) *ChainListener {
	return &ChainListener{nil, make(chan bool), make(map[string]bool), fc, db, maxDiff}
}

func Min(x, y uint64) uint64 {
	if x > y {
		return y
	}
	return x
}

func (l *ChainListener) process(ctx context.Context, start, end uint64) error {
	ee := make([]flow.Event, 0)

	for t := range l.types {
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
		ChainEvent.Trigger(event)
	}

	return nil
}

func (l *ChainListener) handleError(err error) {
	fmt.Println(err)
}

func (l *ChainListener) AddType(t string) {
	l.types[t] = true
}

func (l *ChainListener) RemoveType(t string) {
	delete(l.types, t)
}

func (l *ChainListener) Start() *ChainListener {
	if l.ticker != nil {
		return l
	}

	l.ticker = time.NewTicker(period)

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
				cur, _ := l.fc.GetLatestBlock(ctx, true)
				curHeight := cur.Height
				if curHeight > status.LatestHeight {
					// latestHeight has already been checked, add 1
					start := status.LatestHeight + 1
					end := Min(start+l.maxDiff, curHeight) // Limit maximum end
					if err := l.process(ctx, start, end); err != nil {
						l.handleError(err)
					} else {
						status.LatestHeight = end
						err := l.db.UpdateListenerStatus(status)
						if err != nil {
							l.handleError(err)
						}
					}
				}
			}
		}
	}()

	return l
}

func (l *ChainListener) Stop() {
	l.ticker.Stop()
	l.done <- true
	l.ticker = nil
}
