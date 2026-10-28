package retry

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDoSucceedsFirstTry(t *testing.T) {
	calls := 0
	err := Do(context.Background(), Options{
		Sleep: func(context.Context, time.Duration) error { return nil },
	}, func(context.Context) error {
		calls++
		return nil
	})
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if calls != 1 {
		t.Fatalf("calls = %d, want 1", calls)
	}
}

func TestDoRetriesTransientThenSucceeds(t *testing.T) {
	transient := errors.New("transient")
	calls := 0
	err := Do(context.Background(), Options{
		MaxAttempts: 3,
		BaseDelay:   time.Millisecond,
		MaxDelay:    time.Millisecond,
		RetryIf:     func(err error) bool { return errors.Is(err, transient) },
		Sleep:       func(context.Context, time.Duration) error { return nil },
	}, func(context.Context) error {
		calls++
		if calls < 3 {
			return transient
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if calls != 3 {
		t.Fatalf("calls = %d, want 3", calls)
	}
}

func TestDoNonRetryableError(t *testing.T) {
	permanent := errors.New("permanent")
	calls := 0
	err := Do(context.Background(), Options{
		MaxAttempts: 5,
		RetryIf:     func(err error) bool { return !errors.Is(err, permanent) },
		Sleep:       func(context.Context, time.Duration) error { return nil },
	}, func(context.Context) error {
		calls++
		return permanent
	})
	if !errors.Is(err, permanent) {
		t.Fatalf("err = %v, want permanent", err)
	}
	if calls != 1 {
		t.Fatalf("calls = %d, want 1", calls)
	}
}

func TestDoContextCanceledMidBackoff(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	err := Do(ctx, Options{
		MaxAttempts: 5,
		BaseDelay:   time.Hour,
		MaxDelay:    time.Hour,
		Sleep: func(ctx context.Context, _ time.Duration) error {
			cancel()
			return ctx.Err()
		},
	}, func(context.Context) error {
		calls++
		return errors.New("fail")
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
	if calls != 1 {
		t.Fatalf("calls = %d, want 1", calls)
	}
}

func TestDoHonorsParentDeadline(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	calls := 0
	err := Do(ctx, Options{
		MaxAttempts: 3,
		Sleep:       func(context.Context, time.Duration) error { return nil },
	}, func(context.Context) error {
		calls++
		return errors.New("fail")
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
	if calls != 0 {
		t.Fatalf("calls = %d, want 0", calls)
	}
}

func TestDoOnRetryCalled(t *testing.T) {
	var onRetryAttempts []int
	_ = Do(context.Background(), Options{
		MaxAttempts: 3,
		Sleep:       func(context.Context, time.Duration) error { return nil },
		OnRetry: func(_ context.Context, attempt int, _ error) {
			onRetryAttempts = append(onRetryAttempts, attempt)
		},
	}, func(context.Context) error {
		return errors.New("fail")
	})
	if len(onRetryAttempts) != 2 || onRetryAttempts[0] != 1 || onRetryAttempts[1] != 2 {
		t.Fatalf("onRetryAttempts = %v, want [1 2]", onRetryAttempts)
	}
}

func TestFullJitterBounds(t *testing.T) {
	base := 50 * time.Millisecond
	maxDelay := time.Second
	for attempt := range 10 {
		wantCap := base
		for range attempt {
			if wantCap > maxDelay/2 {
				wantCap = maxDelay
				break
			}
			wantCap *= 2
		}
		if wantCap > maxDelay {
			wantCap = maxDelay
		}
		for range 100 {
			got := FullJitter(base, maxDelay, attempt)
			if got < 0 || got >= wantCap {
				t.Fatalf("FullJitter(attempt=%d) = %v, want in [0, %v)", attempt, got, wantCap)
			}
		}
	}
}

func TestFullJitterZeroInputs(t *testing.T) {
	if got := FullJitter(0, time.Second, 0); got != 0 {
		t.Fatalf("got %v, want 0", got)
	}
	if got := FullJitter(time.Millisecond, 0, 0); got != 0 {
		t.Fatalf("got %v, want 0", got)
	}
}
