package events

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/eqlabs/flow-wallet-service/templates"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"gorm.io/gorm"
)

// TODO: increase once implementation is somewhat complete
const (
	period          = 1 * time.Second
	chanTimeout     = period / 2
	TokensDeposited = "TokensDeposited"
)

type Listener struct {
	ticker  *time.Ticker
	done    chan bool
	types   map[string]bool
	Events  chan []flow.Event
	fc      *client.Client
	db      Store
	maxDiff uint64
}

type ListenerStatus struct {
	gorm.Model
	latestHeight uint64
}

type TimeOutError struct {
	error
}

func NewListener(fc *client.Client, db Store, maxDiff uint64) *Listener {
	return &Listener{nil, make(chan bool), make(map[string]bool), nil, fc, db, maxDiff}
}

func TypeFromToken(t templates.BasicToken, tokenEvent string) string {
	address := strings.TrimPrefix(t.Address, "0x")
	return fmt.Sprintf("A.%s.%s.%s", address, t.Name, tokenEvent)
}

func Min(x, y uint64) uint64 {
	if x > y {
		return y
	}
	return x
}

func (l *Listener) process(ctx context.Context, start, end uint64) error {
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

	select {
	case l.Events <- ee:
		// Sent
		return nil
	case <-time.After(chanTimeout):
		// Timed out while waiting for channel
		return &TimeOutError{}
	}
}

func (l *Listener) handleError(err error) {
	_, ok := err.(*TimeOutError)
	if ok {
		// Ignore timeout errors
		return
	}
	fmt.Println(err)
}

func (l *Listener) AddType(t string) {
	l.types[t] = true
}

func (l *Listener) RemoveType(t string) {
	delete(l.types, t)
}

func (l *Listener) ListenTokenEvent(t templates.BasicToken, tokenEvent string) {
	l.AddType(TypeFromToken(t, tokenEvent))
}

func (l *Listener) RemoveTokenEvent(t templates.BasicToken, tokenEvent string) {
	l.RemoveType(TypeFromToken(t, tokenEvent))
}

func (l *Listener) Start() *Listener {
	if l.ticker != nil {
		return l
	}

	l.ticker = time.NewTicker(period)
	l.Events = make(chan []flow.Event)

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
				if curHeight > status.latestHeight {
					// latestHeight has already been checked, add 1
					start := status.latestHeight + 1
					end := Min(start+l.maxDiff, curHeight) // Limit maximum end
					if err := l.process(ctx, start, end); err != nil {
						l.handleError(err)
					} else {
						status.latestHeight = end
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

func (l *Listener) Stop() {
	l.ticker.Stop()
	l.done <- true
	close(l.Events)
	l.ticker = nil
}
