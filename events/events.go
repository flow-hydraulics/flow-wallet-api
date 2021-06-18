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
	ticker *time.Ticker
	done   chan bool
	types  map[string]bool
	Events chan []flow.Event
	fc     *client.Client
	db     Store
}

type ListenerStatus struct {
	gorm.Model
	latestHeight uint64
}

type TimeOutError struct {
	error
}

func NewListener(fc *client.Client, db Store) *Listener {
	return &Listener{nil, make(chan bool), make(map[string]bool), nil, fc, db}
}

func TypeFromToken(t templates.Token, tokenEvent string) string {
	a := strings.TrimPrefix(t.Address, "0x")
	n := t.CanonName()
	return fmt.Sprintf("A.%s.%s.%s", a, n, tokenEvent)
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

func (l *Listener) ListenTokenEvent(t templates.Token, tokenEvent string) {
	l.AddType(TypeFromToken(t, tokenEvent))
}

func (l *Listener) RemoveTokenEvent(t templates.Token, tokenEvent string) {
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

		// TODO:

		// psiemens:
		//
		// However, if the listener falls behind, it may need to catch up on many
		// blocks at once (e.g. 1000+). The chain can only handle queries for
		// roughly 200 blocks at a time, so you'll need to batch these requests.
		//
		// You could do this by enforcing a maximum value on the block difference
		// (end - start) below. It probably makes to make this maximum a
		// configurable value.

		for {
			select {
			case <-l.done:
				return
			case <-l.ticker.C:
				cur, _ := l.fc.GetLatestBlock(ctx, true)
				end := cur.Height
				if end > status.latestHeight {
					// latestHeight has already been checked, add 1
					if err := l.process(ctx, status.latestHeight+1, end); err != nil {
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
