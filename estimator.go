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
	"hash/maphash"
	"math"
	"sync/atomic"
)

const (
	// DefaultHashes represents the default value for the number of hashes.
	DefaultHashes = 4
	// DefaultSlots represents the default value for the number of slots.
	DefaultSlots = 8192
)

// Estimator is an implementation of the count-min sketch data structure.
//
// An Estimator instance is lock-free, but is safe to use concurrency from
// multiple goroutines.
//
// For more info: https://en.wikipedia.org/wiki/Count%E2%80%93min_sketch
type Estimator[K comparable] struct {
	rows, columns int
	seed1, seed2  maphash.Seed
	data          []atomic.Int64
}

// NewEstimator returns a new Estimator instance using the default hash and slot
// sizes.
func NewEstimator[K comparable]() *Estimator[K] {
	return NewEstimatorWithSize[K](DefaultHashes, DefaultSlots)
}

// NewEstimatorWithSize returns a new Estimator instance using the provided
// number of hashes and slots. This function panics if hashes or slots are less
// than or equal to zero, or hashes is greater than 64. The number of slots is
// rounded up to the nearest power of 2.
func NewEstimatorWithSize[K comparable](hashes, slots int) *Estimator[K] {
	if hashes <= 0 {
		panic("limits: hashes must be greater than 0")
	}
	if hashes > 64 {
		panic("limits: hashes must be less than or equal to 64")
	}
	if slots <= 0 {
		panic("limits: slots must be greater than 0")
	}

	slots = int(roundUpPow2(uint64(slots)))
	data := make([]atomic.Int64, hashes*slots)
	seed1 := maphash.MakeSeed()
	seed2 := maphash.MakeSeed()

	return &Estimator[K]{hashes, slots, seed1, seed2, data}
}

// Get returns the estimated count for the provided key.
func (e *Estimator[K]) Get(key K) int64 {
	h1, h2 := maphash.Comparable(e.seed1, key), maphash.Comparable(e.seed2, key)

	minimum := int64(math.MaxInt64)
	for i := 0; i < e.rows; i++ {
		index := e.index(h1, h2, i)
		count := e.data[i*e.columns+index].Load()
		minimum = min(minimum, count)
	}
	return minimum
}

// Incr is the equivalent of calling `IncrN(key, 1)`.
func (e *Estimator[K]) Incr(key K) int64 {
	return e.IncrN(key, 1)
}

// IncrN increments the count by 'n' for the provided key, returning the
// estimated total count.
func (e *Estimator[K]) IncrN(key K, n int64) int64 {
	h1, h2 := maphash.Comparable(e.seed1, key), maphash.Comparable(e.seed2, key)

	// First pass: get rows that match the minimum.
	var offsets [64]int
	var mask uint64
	minimum := int64(math.MaxInt64)
	for i := range e.rows {
		offsets[i] = i*e.columns + e.index(h1, h2, i)
		count := e.data[offsets[i]].Load()
		if count < minimum {
			minimum = count
			mask = 1 << uint64(i)
		} else if count == minimum {
			mask |= 1 << uint64(i)
		}
	}

	// Second pass: increment only the rows that matched the minimum,
	// but take the minimum across ALL rows.
	minimum = int64(math.MaxInt64)
	for i := range e.rows {
		var count int64
		if (mask>>uint(i))&1 == 1 {
			count = e.data[offsets[i]].Add(n)
		} else {
			count = e.data[offsets[i]].Load()
		}
		minimum = min(minimum, count)
	}

	return minimum
}

// Reset clears the Estimator, returning all counts to 0.
func (e *Estimator[K]) Reset() {
	for i := range e.data {
		e.data[i].Store(0)
	}
}

func (e *Estimator[K]) index(h1, h2 uint64, row int) int {
	return int((h1 + uint64(row)*h2) & uint64(e.columns-1))
}

func roundUpPow2(x uint64) uint64 {
	if x <= 1 {
		return 1
	}
	x--
	x |= x >> 1
	x |= x >> 2
	x |= x >> 4
	x |= x >> 8
	x |= x >> 16
	x |= x >> 32
	return x + 1
}
