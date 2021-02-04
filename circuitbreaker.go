package circuitbreaker

import (
	"context"
	"time"
)

var getNow = time.Now

type State int8

const (
	Closed State = iota
	HalfOpen
	Open
)

type CircuitBreaker struct {
	clock Clock

	state State

	errorCount int

	lastStateChanged time.Time
	lastErr          error

	threshold    int
	openDuration time.Duration
}

type Clock interface {
	Now() time.Time
}

type systemClock struct{}

func (systemClock) Now() time.Time {
	return time.Now()
}

func New(threshold int, openDuration time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		threshold:    threshold,
		openDuration: openDuration,

		clock: systemClock{},
	}
}

type Action interface {
	Do(context.Context) error
}

type ActionFunc func(context.Context) error

func (action ActionFunc) Do(ctx context.Context) error {
	return action(ctx)
}

func (cb *CircuitBreaker) IsOpen() bool {
	return cb.state != Closed
}

func (cb *CircuitBreaker) Do(ctx context.Context, action Action) {
	now := cb.clock.Now()

	if cb.IsOpen() {
		if cb.lastStateChanged.Add(cb.openDuration).After(now) {
			return
		}

		cb.state = HalfOpen
	}

	if err := action.Do(ctx); err != nil {
		cb.errorCount++

		if cb.threshold <= cb.errorCount {
			cb.state = Open
			cb.lastStateChanged = now
			cb.lastErr = err
			cb.errorCount = 0
		}
	}
}

func (cb *CircuitBreaker) LastStateChanged() time.Time {
	return cb.lastStateChanged
}

func (cb *CircuitBreaker) LastErr() error {
	return cb.lastErr
}

func (cb *CircuitBreaker) UseClock(clock Clock) {
	cb.clock = clock
}
