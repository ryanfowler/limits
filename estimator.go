package limits

import (
	"hash/maphash"
	"math"
	"sync/atomic"
)

type hashSlots struct {
	seed  maphash.Seed
	slots []atomic.Int64
}

// Estimator is an implementation of the count-min sketch data structure.
//
// An Estimator instance is lock-free, but is safe to use concurrency from
// multiple goroutines.
//
// For more info: https://en.wikipedia.org/wiki/Count%E2%80%93min_sketch
type Estimator struct {
	data []hashSlots
}

// NewEstimator returns a new Estimator instance using the provided number of
// hashes and slots.
func NewEstimator(hashes, slots int) Estimator {
	if hashes <= 0 {
		panic("limits: hashes must be greater than 0")
	}
	if slots <= 0 {
		panic("limits: slots must be greater than 0")
	}

	data := make([]hashSlots, hashes)
	for i := 0; i < len(data); i++ {
		data[i] = hashSlots{
			seed:  maphash.MakeSeed(),
			slots: make([]atomic.Int64, slots),
		}
	}

	return Estimator{data}
}

// GetString returns the estimated count for the provided key.
func (e Estimator) GetString(key string) int64 {
	return e.get(func(seed maphash.Seed) uint64 { return maphash.String(seed, key) })
}

// GetBytes returns the estimated count for the provided key.
func (e Estimator) GetBytes(key []byte) int64 {
	return e.get(func(seed maphash.Seed) uint64 { return maphash.Bytes(seed, key) })
}

// IncrString is the equivalent of calling `IncrNString(key, 1)`.
func (e Estimator) IncrString(key string) int64 {
	return e.IncrNString(key, 1)
}

// IncrNString increments the count by 'n' for the provided key, returning the
// estimated total count.
func (e Estimator) IncrNString(key string, n int64) int64 {
	return e.incr(n, func(seed maphash.Seed) uint64 { return maphash.String(seed, key) })
}

// IncrBytes is the equivalent of calling `IncrNBytes(key, 1)`.
func (e Estimator) IncrBytes(key []byte) int64 {
	return e.IncrNBytes(key, 1)
}

// IncrNBytes increments the count by 'n' for the provided key, returning the
// estimated total count.
func (e Estimator) IncrNBytes(key []byte, n int64) int64 {
	return e.incr(n, func(seed maphash.Seed) uint64 { return maphash.Bytes(seed, key) })
}

// Reset clears the Estimator, returning all counts to 0.
func (e Estimator) Reset() {
	for _, hs := range e.data {
		for i := 0; i < len(hs.slots); i++ {
			hs.slots[i].Store(0)
		}
	}
}

func (e Estimator) get(fn func(maphash.Seed) uint64) int64 {
	var minimum int64 = math.MaxInt64
	for _, hs := range e.data {
		hash := fn(hs.seed)
		count := hs.slots[int(hash%uint64(len(hs.slots)))].Load()
		minimum = min(minimum, count)
	}
	return minimum
}

func (e Estimator) incr(n int64, fn func(maphash.Seed) uint64) int64 {
	var minimum int64 = math.MaxInt64
	for _, hs := range e.data {
		hash := fn(hs.seed)
		count := hs.slots[int(hash%uint64(len(hs.slots)))].Add(n)
		minimum = min(minimum, count)
	}
	return minimum
}
