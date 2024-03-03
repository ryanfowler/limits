// Copyright 2024 Ryan Fowler
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package limits

import (
	"sync/atomic"
	"time"
)

const (
	// DefaultRateHashes represents the default value for the number of hashes.
	DefaultRateHashes = 4
	// DefaultRateSlots represents the default value for the number of slots.
	DefaultRateSlots = 1024
)

// Rate is a probabilistic rate estimator over a given interval. Internally, it
// uses multiple `Estimator`s to track the number of events per key.
//
// A Rate instance is lock-free, but is safe to use concurrency from multiple
// goroutines.
type Rate[K Key] struct {
	red, blue       Estimator[K]
	isRed           atomic.Bool
	start           time.Time
	resetIntervalMs int64
	lastResetTime   atomic.Int64
}

// NewRate returns a new Rate instance using the provided interval and default
// sizes. NewRate panics if interval is smaller than 1 millisecond.
func NewRate[K Key](interval time.Duration) *Rate[K] {
	return NewRateWithSize[K](interval, DefaultRateHashes, DefaultRateSlots)
}

// NewRateWithSize returns a new Rate instance using the provided interval and
// hash/slot sizes. NewRateWithSize panics if interval is smaller than 1
// millisecond.
func NewRateWithSize[K Key](interval time.Duration, hashes, slots int) *Rate[K] {
	if interval < time.Millisecond {
		panic("limits: interval must be 1 millisecond or greater")
	}
	return &Rate[K]{
		red:             NewEstimatorWithSize[K](hashes, slots),
		blue:            NewEstimatorWithSize[K](hashes, slots),
		isRed:           atomic.Bool{},
		start:           time.Now().UTC(),
		resetIntervalMs: interval.Milliseconds(),
		lastResetTime:   atomic.Int64{},
	}
}

// Get returns the total estimated number of events in the previous interval.
func (r *Rate[K]) Get(key K) int64 {
	pastMs := r.maybeReset()
	if pastMs >= 2*r.resetIntervalMs {
		return 0
	}
	return r.getEstimator(!r.isRed.Load()).Get(key)
}

// Observe is the equivalent of calling `ObserveN(key, 1)`.
func (r *Rate[K]) Observe(key K) int64 {
	return r.ObserveN(key, 1)
}

// ObserveN records 'n' events for the provided key, returning the total
// estimated number of events in the current interval.
func (r *Rate[K]) ObserveN(key K, n int64) int64 {
	r.maybeReset()
	return r.getEstimator(r.isRed.Load()).IncrN(key, n)
}

func (r *Rate[K]) maybeReset() int64 {
	now := time.Since(r.start).Milliseconds()
	lastReset := r.lastResetTime.Load()
	pastMs := now - lastReset

	if pastMs < r.resetIntervalMs {
		return pastMs
	}

	if r.lastResetTime.CompareAndSwap(lastReset, now) {
		isRed := r.isRed.Load()
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

func (r *Rate[K]) getEstimator(isRed bool) Estimator[K] {
	if isRed {
		return r.red
	}
	return r.blue
}

func abs(n int64) int64 {
	mask := n >> 63
	return (mask + n) ^ mask
}
