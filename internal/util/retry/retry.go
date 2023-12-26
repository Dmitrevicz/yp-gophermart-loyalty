// Package retry.
package retry

import (
	"errors"
	"time"

	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/logger"
	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/model"
	"go.uber.org/zap"
)

type Retrier struct {
	interval time.Duration
	retries  int
	retryAny bool
	infinite bool
}

type RetrierOptions struct {
	Interval time.Duration
	Retries  int
	RetryAny bool // retry any error (and not retriable ones only)
	Infinite bool // ignore Retries number and run infinitely until success
}

func NewRetrier(opt RetrierOptions) *Retrier {
	if opt.Interval <= 0 {
		opt.Interval = time.Millisecond * 500
	}

	if opt.Retries < 0 {
		opt.Retries = 0
	}

	r := Retrier{
		interval: opt.Interval,
		retries:  opt.Retries,
		retryAny: opt.RetryAny,
		infinite: opt.Infinite,
	}

	return &r
}

const progressionLimit = 5

// progression decides what duration till next attempt should be waited
func (r *Retrier) progression(currentAttempt int) {
	if currentAttempt <= 1 || currentAttempt > progressionLimit {
		// interval duration will be increased at max 5 times
		// to prevent potentially endless wait
		return
	}

	// TODO: add random jitter as told somewhere in best practices
	// https://aws.amazon.com/blogs/architecture/exponential-backoff-and-jitter/
	r.interval = (r.interval + time.Millisecond*500) * 2
	// 0.5, 2.0, 5.0, 11.0, 23.0
}

// Do does a retry of f().
func (r *Retrier) Do(action string, f func() error) (err error) {
	var retriable model.RetriableError

	i := 0
	for {
		if !r.infinite && i > r.retries {
			break
		}

		if i > 0 {
			time.Sleep(r.interval)
			r.progression(i)
			logger.Log.Info("retrying...", zap.String("action", action),
				zap.Int("attempt", i),
				zap.Error(err),
			)
		}

		err = f()
		if err == nil || (!r.retryAny && !errors.As(err, &retriable)) {
			break // no need to try again
		}

		i++

		// handle int overflow because loop can be set up to run infinitely
		if i < 0 {
			i = progressionLimit + 1
		}
	}

	return err
}
