package circuitbreaker_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/nasa9084/go-circuitbreaker"
)

type fakeClock struct {
	time.Time
}

func (c *fakeClock) Now() time.Time              { return c.Time }
func (c *fakeClock) FastForward(d time.Duration) { c.Time = c.Time.Add(d) }

func TestCircuitBreaker(t *testing.T) {
	clock := &fakeClock{
		Time: time.Date(2021, time.February, 4, 23, 26, 0, 0, time.UTC),
	}
	cb := circuitbreaker.New(3, 3*time.Second)
	cb.UseClock(clock)
	errFail := errors.New("fail")
	failAction := circuitbreaker.ActionFunc(func(ctx context.Context) error {
		return errFail
	})
	ctx := context.Background()

	// First time attempt: not opened
	if cb.Do(ctx, failAction); cb.IsOpen() {
		t.Error("expected to be closed")
		return
	}

	// Second time attempt: not opened
	if cb.Do(ctx, failAction); cb.IsOpen() {
		t.Error("expected to be closed")
		return
	}

	// Third time attempt: should opened
	if cb.Do(ctx, failAction); !cb.IsOpen() {
		t.Error("expected to be opened")
		return
	}

	if cb.LastStateChanged() != clock.Time {
		t.Errorf("unexpected last state changed time: %s != %s",
			cb.LastStateChanged().Format(time.RFC3339),
			clock.Time.Format(time.RFC3339),
		)
		return
	}

	if cb.LastErr() != errFail {
		t.Errorf("unexpected last error: %v != %v", cb.LastErr(), errFail)
		return
	}

	shouldNotAction := circuitbreaker.ActionFunc(func(ctx context.Context) error {
		t.Error("expected not to be executed")
		return nil
	})

	cb.Do(ctx, shouldNotAction)

	clock.FastForward(1 * time.Second)
	cb.Do(ctx, shouldNotAction)

	clock.FastForward(1 * time.Second)
	cb.Do(ctx, shouldNotAction)

	clock.FastForward(1 * time.Second)

	var called bool
	shouldAction := circuitbreaker.ActionFunc(func(ctx context.Context) error {
		called = true
		return nil
	})

	cb.Do(ctx, shouldAction)

	if !called {
		t.Error("circuit breaker should be closed and action should be called but not")
		return
	}
}
