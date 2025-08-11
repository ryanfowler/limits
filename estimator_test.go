package limits

import "testing"

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
