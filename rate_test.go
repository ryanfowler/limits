package limits

import (
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestRateBasic(t *testing.T) {
	key := "1"
	r := NewRate[string](10 * time.Millisecond)

	v := r.Observe(key)
	assertIntsEqual(t, v, 1)
	v = r.ObserveN(key, 2)
	assertIntsEqual(t, v, 3)
	v = r.Get(key)
	assertIntsEqual(t, v, 0)

	time.Sleep(11 * time.Millisecond)
	v = r.Observe(key)
	assertIntsEqual(t, v, 1)
	v = r.Get(key)
	assertIntsEqual(t, v, 3)

	time.Sleep(11 * time.Millisecond)
	v = r.Get(key)
	assertIntsEqual(t, v, 1)

	time.Sleep(11 * time.Millisecond)
	v = r.Get(key)
	assertIntsEqual(t, v, 0)
}

func TestRateConcurrency(t *testing.T) {
	r := NewRate[string](time.Second)

	var wg sync.WaitGroup
	for i := range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			key := strconv.Itoa(i + 1)
			for j := range 1000 {
				v := r.Observe(key)
				if v != int64(j+1) {
					panic(fmt.Sprintf("unexpected value: %d", v))
				}
			}
		}()
	}
	wg.Wait()

	for i := range 10 {
		key := strconv.Itoa(i + 1)
		if v := r.Observe(key); v != 1001 {
			t.Fatalf("unexpected value: %v", v)
		}
	}
}

func TestRateConcurrentReset(t *testing.T) {
	const interval = 200 * time.Millisecond

	r := NewRate[string](interval)

	// Fill the current interval with observations.
	for i := range 100 {
		r.Observe(strconv.Itoa(i))
	}

	// Sleep past one interval so the data moves to the inactive estimator.
	time.Sleep(interval + 10*time.Millisecond)

	// Verify the data is still accessible before concurrent access.
	if v := r.Get("0"); v != 1 {
		t.Fatalf("precondition failed: Get(\"0\") = %d, want 1", v)
	}

	// Race many goroutines on Get. After the initial maybeReset above,
	// subsequent calls within this interval should not trigger another
	// reset and should all see the data.
	var wg sync.WaitGroup
	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for _, key := range []string{"0", "1", "2", "3", "4"} {
				v := r.Get(key)
				if v == 0 {
					t.Errorf("Get(%q) returned 0 after single interval; expected > 0", key)
					return
				}
			}
		}()
	}
	wg.Wait()
}

func BenchmarkRateGetString(b *testing.B) {
	r := NewRate[string](time.Second)
	for b.Loop() {
		r.Get("a")
	}
}

func BenchmarkRateObserveString(b *testing.B) {
	r := NewRate[string](time.Second)
	for b.Loop() {
		r.Observe("a")
	}
}
