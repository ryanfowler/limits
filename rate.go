package limits

import (
	"sync/atomic"
	"time"
)

const (
	// DefaultHashes represents the default value for the number of hashes.
	DefaultHashes = 4
	// DefaultSlots represents the default value for the number of slots.
	DefaultSlots = 1024
)

// Rate is a probabilistic rate estimator over a given interval. Internally, it
// uses multiple `Estimator`s to track the number of events per key.
//
// A Rate instance is lock-free, but is safe to use concurrency from multiple
// goroutines.
type Rate struct {
	red, blue       Estimator
	isRed           atomic.Bool
	start           time.Time
	resetIntervalMs int64
	lastResetTime   atomic.Int64
}

// NewRate returns a new Rate instance using the provided interval and default
// sizes. NewRate panics if interval is smaller than 1 millisecond.
func NewRate(interval time.Duration) *Rate {
	return NewRateWithSize(interval, DefaultHashes, DefaultSlots)
}

// NewRateWithSize returns a new Rate instance using the provided interval and
// hash/slot sizes. NewRateWithSize panics if interval is smaller than 1
// millisecond.
func NewRateWithSize(interval time.Duration, hashes, slots int) *Rate {
	if interval < time.Millisecond {
		panic("limits: interval must be 1 millisecond or greater")
	}
	return &Rate{
		red:             NewEstimator(hashes, slots),
		blue:            NewEstimator(hashes, slots),
		isRed:           atomic.Bool{},
		start:           time.Now().UTC(),
		resetIntervalMs: interval.Milliseconds(),
		lastResetTime:   atomic.Int64{},
	}
}

// PerSecondString returns the estimated rate per second for the provided key
// based on the previous interval.
func (r *Rate) PerSecondString(key string) float64 {
	return r.perSecond(func(e Estimator) int64 { return e.GetString(key) })
}

// PerSecondBytes returns the estimated rate per second for the provided key
// based on the previous interval.
func (r *Rate) PerSecondBytes(key []byte) float64 {
	return r.perSecond(func(e Estimator) int64 { return e.GetBytes(key) })
}

// ObserveString is the equivalent of calling `ObserveNString(key, 1)`.
func (r *Rate) ObserveString(key string) int64 {
	return r.ObserveNString(key, 1)
}

// ObserveNString records 'n' events for the provided key, returning the total
// estimated number of events for the current interval.
func (r *Rate) ObserveNString(key string, n int64) int64 {
	r.maybeReset()
	return r.getEstimator(r.isRed.Load()).IncrNString(key, n)
}

// ObserveBytes is the equivalent of calling `ObserveNBytes(key, 1)`.
func (r *Rate) ObserveBytes(key []byte) int64 {
	return r.ObserveNBytes(key, 1)
}

// ObserveNBytes records 'n' events for the provided key, returning the total
// estimated number of events for the current interval.
func (r *Rate) ObserveNBytes(key []byte, n int64) int64 {
	r.maybeReset()
	return r.getEstimator(r.isRed.Load()).IncrNBytes(key, n)
}

func (r *Rate) maybeReset() int64 {
	now := time.Since(r.start).Milliseconds()
	lastReset := r.lastResetTime.Load()
	pastMs := now - lastReset

	if pastMs < r.resetIntervalMs {
		return pastMs
	}

	isRed := r.isRed.Load()
	if r.lastResetTime.CompareAndSwap(lastReset, now) {
		r.getEstimator(!isRed).Reset()
		r.isRed.Store(!isRed)

		// If the current time is beyond 2 intervals, we should reset
		// the previous Estimator as well.
		if pastMs >= 2*r.resetIntervalMs {
			r.getEstimator(isRed).Reset()
		}
	}

	return pastMs
}

func (r *Rate) getEstimator(isRed bool) Estimator {
	if isRed {
		return r.red
	}
	return r.blue
}

func (r *Rate) perSecond(fn func(Estimator) int64) float64 {
	pastMs := r.maybeReset()
	if pastMs >= 2*r.resetIntervalMs {
		return 0.0
	}

	return 1000.0 * float64(fn(r.getEstimator(!r.isRed.Load()))) / float64(r.resetIntervalMs)
}
