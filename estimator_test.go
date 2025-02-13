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

func BenchmarkEstimatorGetString(b *testing.B) {
	e := NewEstimatorWithSize[string](4, 1024)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Get("a")
	}
}

func BenchmarkEstimatorIncrString(b *testing.B) {
	e := NewEstimatorWithSize[string](4, 1024)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Incr("a")
	}
}

func BenchmarkEstimatorReset(b *testing.B) {
	e := NewEstimatorWithSize[string](4, 1024)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Reset()
	}
}
