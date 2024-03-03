package limits

import (
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestRateBasic(t *testing.T) {
	key := "1"
	r := NewRate[string](10 * time.Millisecond)

	v := r.Observe(key)
	assertFloatBetween(t, v, 0.000001, 1.0)
	v = r.ObserveN(key, 2)
	assertFloatBetween(t, v, 0.000001, 3.0)
	v = r.Get(key)
	assertFloatBetween(t, v, 0.000001, 3.0)

	time.Sleep(11 * time.Millisecond)
	v = r.Observe(key)
	assertFloatBetween(t, v, 1.0, 4.0)
	v = r.Get(key)
	assertFloatBetween(t, v, 1.0, 4.0)

	time.Sleep(11 * time.Millisecond)
	v = r.Get(key)
	assertFloatBetween(t, v, 0.000001, 1.0)

	time.Sleep(11 * time.Millisecond)
	v = r.Get(key)
	assertFloatBetween(t, v, -0.01, 0.01)
}

func TestRateConcurrency(t *testing.T) {
	r := NewRate[string](time.Second)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			key := strconv.Itoa(i + 1)
			for j := 0; j < 1000; j++ {
				r.Observe(key)
			}
		}()
	}
	wg.Wait()

	for i := 0; i < 10; i++ {
		key := strconv.Itoa(i + 1)
		if v := r.Get(key); v < 100.0 || v > 1000.0 {
			t.Fatalf("unexpected value: %v", v)
		}
	}
}

func assertFloatBetween(t *testing.T, got, min, max float64) {
	t.Helper()

	if got > max || got < min {
		t.Fatalf("expected between'%f' and '%f', got '%f'", min, max, got)
	}
}

func BenchmarkGetString(b *testing.B) {
	r := NewRate[string](time.Second)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Get("a")
	}
}

func BenchmarkGetBytes(b *testing.B) {
	r := NewRate[[]byte](time.Second)
	key := []byte("a")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Get(key)
	}
}

func BenchmarkRateObserveString(b *testing.B) {
	r := NewRate[string](time.Second)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Observe("a")
	}
}

func BenchmarkRateObserveBytes(b *testing.B) {
	r := NewRate[[]byte](time.Second)
	key := []byte("a")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Observe(key)
	}
}
