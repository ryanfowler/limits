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
	data []hashSlots
}

type hashSlots struct {
	seed  maphash.Seed
	slots []atomic.Int64
}

// NewEstimator returns a new Estimator instance using the default hash and slot
// sizes.
func NewEstimator[K comparable]() Estimator[K] {
	return NewEstimatorWithSize[K](DefaultHashes, DefaultSlots)
}

// NewEstimatorWithSize returns a new Estimator instance using the provided
// number of hashes and slots. This function panics if hashes or slots are less
// than or equal to zero.
func NewEstimatorWithSize[K comparable](hashes, slots int) Estimator[K] {
	if hashes <= 0 {
		panic("limits: hashes must be greater than 0")
	}
	if slots <= 0 {
		panic("limits: slots must be greater than 0")
	}

	data := make([]hashSlots, hashes)
	for i := range data {
		data[i] = hashSlots{
			seed:  maphash.MakeSeed(),
			slots: make([]atomic.Int64, slots),
		}
	}

	return Estimator[K]{data}
}

// Get returns the estimated count for the provided key.
func (e Estimator[K]) Get(key K) int64 {
	var minimum int64 = math.MaxInt64
	for _, hs := range e.data {
		hash := maphash.Comparable(hs.seed, key)
		count := hs.slots[int(hash%uint64(len(hs.slots)))].Load()
		minimum = min(minimum, count)
	}
	return minimum
}

// Incr is the equivalent of calling `IncrN(key, 1)`.
func (e Estimator[K]) Incr(key K) int64 {
	return e.IncrN(key, 1)
}

// IncrN increments the count by 'n' for the provided key, returning the
// estimated total count.
func (e Estimator[K]) IncrN(key K, n int64) int64 {
	var minimum int64 = math.MaxInt64
	for _, hs := range e.data {
		hash := maphash.Comparable(hs.seed, key)
		count := hs.slots[int(hash%uint64(len(hs.slots)))].Add(n)
		minimum = min(minimum, count)
	}
	return minimum
}

// Reset clears the Estimator, returning all counts to 0.
func (e Estimator[K]) Reset() {
	for _, hs := range e.data {
		for i := range hs.slots {
			hs.slots[i].Store(0)
		}
	}
}
