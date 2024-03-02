package limits

import (
	"math"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestRateBasic(t *testing.T) {
	key := "1"
	r := NewRate(10 * time.Millisecond)

	v := r.ObserveString(key)
	assertIntsEqual(t, 1, v)
	v = r.ObserveNString(key, 2)
	assertIntsEqual(t, 3, v)
	f := r.PerSecondString(key)
	assertFloatsEqual(t, 0.0, f)

	time.Sleep(11 * time.Millisecond)
	v = r.ObserveString(key)
	assertIntsEqual(t, 1, v)
	f = r.PerSecondString(key)
	assertFloatsEqual(t, 300.0, f)

	time.Sleep(11 * time.Millisecond)
	f = r.PerSecondString(key)
	assertFloatsEqual(t, 100.0, f)

	time.Sleep(11 * time.Millisecond)
	f = r.PerSecondString(key)
	assertFloatsEqual(t, 0.0, f)
}

func TestRateConcurrency(t *testing.T) {
	r := NewRate(time.Second)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			key := strconv.Itoa(i + 1)
			for j := 0; j < 1000; j++ {
				r.ObserveString(key)
			}
		}()
	}
	wg.Wait()

	for i := 0; i < 10; i++ {
		key := strconv.Itoa(i + 1)
		if v := r.ObserveString(key); v != 1001 {
			t.Fatalf("unexpected value: %d", v)
		}
	}
}

func assertIntsEqual(t *testing.T, exp, got int64) {
	t.Helper()

	if exp != got {
		t.Fatalf("expected '%d', got '%d'", exp, got)
	}
}

func assertFloatsEqual(t *testing.T, exp, got float64) {
	t.Helper()

	const threshold = 1e-9
	if math.Abs(exp-got) > threshold {
		t.Fatalf("expected '%f', got '%f'", exp, got)
	}
}

func BenchmarkRatePerSecondString(b *testing.B) {
	r := NewRate(time.Second)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.PerSecondString("a")
	}
}

func BenchmarkRatePerSecondBytes(b *testing.B) {
	r := NewRate(time.Second)
	key := []byte("a")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.PerSecondBytes(key)
	}
}

func BenchmarkRateObserveString(b *testing.B) {
	r := NewRate(time.Second)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ObserveString("a")
	}
}

func BenchmarkRateObserveBytes(b *testing.B) {
	r := NewRate(time.Second)
	key := []byte("a")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ObserveBytes(key)
	}
}
