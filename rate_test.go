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
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			key := strconv.Itoa(i + 1)
			for j := 0; j < 1000; j++ {
				v := r.Observe(key)
				if v != int64(j+1) {
					panic(fmt.Sprintf("unexpected value: %d", v))
				}
			}
		}()
	}
	wg.Wait()

	for i := 0; i < 10; i++ {
		key := strconv.Itoa(i + 1)
		if v := r.Observe(key); v != 1001 {
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
