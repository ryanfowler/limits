package limits

import (
	"sync"
	"testing"
)

func TestEstimator(t *testing.T) {
	const (
		key1 = "one"
		key2 = "two"
	)
	est := NewEstimator[string]()
	v := est.Incr(key1)
	assertIntsEqual(t, v, 1)
	v = est.Incr(key2)
	assertIntsEqual(t, v, 1)
	v = est.IncrN(key1, 2)
	assertIntsEqual(t, v, 3)
	v = est.IncrN(key2, 3)
	assertIntsEqual(t, v, 4)

	v = est.IncrN(key1, -1)
	assertIntsEqual(t, v, 2)
	v = est.IncrN(key2, -1)
	assertIntsEqual(t, v, 3)

	v = est.Get(key1)
	assertIntsEqual(t, v, 2)
	v = est.Get(key2)
	assertIntsEqual(t, v, 3)

	est.Reset()
	v = est.Get(key1)
	assertIntsEqual(t, v, 0)
	v = est.Get(key2)
	assertIntsEqual(t, v, 0)
}

func TestEstimatorIncrNWithDivergentRows(t *testing.T) {
	// With one slot per row, set up the selected rows with different counts.
	est := NewEstimatorWithSize[string](3, 1)
	est.data[0].Store(0)
	est.data[1].Store(1)
	est.data[2].Store(10)

	if got := est.IncrN("key", 5); got < 5 {
		t.Fatalf("IncrN returned %d, want at least 5", got)
	}
	if got := est.Get("key"); got < 5 {
		t.Fatalf("Get returned %d, want at least 5", got)
	}
}

func TestEstimatorConcurrentIncr(t *testing.T) {
	const (
		workers      = 32
		observations = 1000
	)

	est := NewEstimatorWithSize[string](4, 1)
	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range observations {
				est.Incr("key")
			}
		}()
	}
	wg.Wait()

	want := int64(workers * observations)
	if got := est.Get("key"); got < want {
		t.Fatalf("Get returned %d, want at least %d", got, want)
	}
}

func TestIncrNReturnValue(t *testing.T) {
	// Use a small estimator where hash collisions make rows diverge.
	est := NewEstimatorWithSize[string](4, 16)

	// Increment many distinct keys to create varying row counts, then
	// verify that IncrN always returns the same value as an immediate Get.
	keys := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
	for _, key := range keys {
		for n := int64(1); n <= 5; n++ {
			got := est.IncrN(key, n)
			want := est.Get(key)
			if got != want {
				t.Fatalf("IncrN(%q, %d) = %d, but Get(%q) = %d", key, n, got, key, want)
			}
		}
	}
}

func assertIntsEqual(t *testing.T, got, exp int64) {
	t.Helper()

	if exp != got {
		t.Fatalf("expected '%d', got '%d'", exp, got)
	}
}

func BenchmarkEstimatorGetStringSmall(b *testing.B) {
	e := NewEstimatorWithSize[string](4, 1024)
	for b.Loop() {
		e.Get("a")
	}
}

func BenchmarkEstimatorIncrStringSmall(b *testing.B) {
	e := NewEstimatorWithSize[string](4, 1024)
	for b.Loop() {
		e.Incr("a")
	}
}

func BenchmarkEstimatorResetSmall(b *testing.B) {
	e := NewEstimatorWithSize[string](4, 1024)
	for b.Loop() {
		e.Reset()
	}
}

func BenchmarkEstimatorGetStringLarge(b *testing.B) {
	e := NewEstimatorWithSize[string](8, 8192)
	for b.Loop() {
		e.Get("a")
	}
}

func BenchmarkEstimatorIncrStringLarge(b *testing.B) {
	e := NewEstimatorWithSize[string](8, 8192)
	for b.Loop() {
		e.Incr("a")
	}
}

func BenchmarkEstimatorResetLarge(b *testing.B) {
	e := NewEstimatorWithSize[string](8, 8192)
	for b.Loop() {
		e.Reset()
	}
}
