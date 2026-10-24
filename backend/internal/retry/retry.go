package retry

import (
	"context"
	"math/rand/v2"
	"time"
)

const (
	DefaultMaxAttempts = 3
	DefaultBaseDelay   = 50 * time.Millisecond
	DefaultMaxDelay    = time.Second
)

// Options configures [Do].
type Options struct {
	// MaxAttempts is the maximum number of retry attempts.
	// Default is [defaultMaxAttempts].
	MaxAttempts int

	// BaseDelay is the initial delay before the first retry.
	// Default is [defaultBaseDelay].
	BaseDelay time.Duration

	// MaxDelay is the maximum backoff delay between retries.
	// Default is [defaultMaxDelay].
	MaxDelay time.Duration

	// RetryIf returns true when err should be retried. If nil, all errors are retried.
	RetryIf func(error) bool

	// Sleep waits for d or until ctx is done. Injectable for tests; default uses time.Timer.
	Sleep func(ctx context.Context, d time.Duration) error

	// OnRetry is called after a failed attempt that will be retried, before sleeping.
	// attempt is the 1-based index of the attempt that just failed.
	OnRetry func(ctx context.Context, attempt int, err error)
}

// Do runs fn with timeout-aware retries using exponential backoff + full jitter.
// It never extends the parent deadline; retries stop when ctx is done.
func Do(ctx context.Context, opts Options, fn func(context.Context) error) error {
	opts = opts.withDefaults()

	var lastErr error
	for attempt := 1; attempt <= opts.MaxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		lastErr = fn(ctx)
		if lastErr == nil {
			return nil
		}

		if attempt == opts.MaxAttempts || !opts.RetryIf(lastErr) {
			return lastErr
		}

		if opts.OnRetry != nil {
			opts.OnRetry(ctx, attempt, lastErr)
		}

		delay := FullJitter(opts.BaseDelay, opts.MaxDelay, attempt-1)
		if err := opts.Sleep(ctx, delay); err != nil {
			return err
		}
	}
	return lastErr
}

func (o Options) withDefaults() Options {
	if o.MaxAttempts <= 0 {
		o.MaxAttempts = DefaultMaxAttempts
	}
	if o.BaseDelay <= 0 {
		o.BaseDelay = DefaultBaseDelay
	}
	if o.MaxDelay <= 0 {
		o.MaxDelay = DefaultMaxDelay
	}
	if o.RetryIf == nil {
		o.RetryIf = func(error) bool { return true }
	}
	if o.Sleep == nil {
		o.Sleep = sleep
	}
	return o
}

// FullJitter returns a delay in [0, min(maxDelay, base*2^attempt)).
// AWS-style full jitter: sleep = random(0, min(cap, base*2^attempt)).
func FullJitter(base, maxDelay time.Duration, attempt int) time.Duration {
	if base <= 0 || maxDelay <= 0 {
		return 0
	}
	if attempt < 0 {
		attempt = 0
	}

	capDelay := base
	for i := 0; i < attempt; i++ {
		nextCapDelay := capDelay * 2
		if nextCapDelay > maxDelay {
			capDelay = maxDelay
			break
		}
		capDelay = nextCapDelay
	}
	if capDelay <= 0 {
		return 0
	}
	return time.Duration(rand.Int64N(int64(capDelay)))
}

func sleep(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return ctx.Err()
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
