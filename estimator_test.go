package limits

import "testing"

func BenchmarkEstimatorGetString(b *testing.B) {
	e := NewEstimator(4, 1024)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.GetString("a")
	}
}

func BenchmarkEstimatorGetBytes(b *testing.B) {
	e := NewEstimator(4, 1024)
	key := []byte("a")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.GetBytes(key)
	}
}

func BenchmarkEstimatorIncrString(b *testing.B) {
	e := NewEstimator(4, 1024)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.IncrString("a")
	}
}

func BenchmarkEstimatorIncrBytes(b *testing.B) {
	e := NewEstimator(4, 1024)
	key := []byte("a")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.IncrBytes(key)
	}
}

func BenchmarkEstimatorReset(b *testing.B) {
	e := NewEstimator(4, 1024)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Reset()
	}
}
