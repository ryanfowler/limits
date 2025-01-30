package limits

// type Inflight struct {
// 	est Estimator
// }

// func NewInflight() Inflight {
// 	return NewInflightWithSize(DefaultHashes, DefaultSlots)
// }

// func NewInflightWithSize(hashes, slots int) Inflight {
// 	return Inflight{NewEstimator(hashes, slots)}
// }

// func (i Inflight) GetString(key string) int64 {
// 	return i.est.GetString(key)
// }

// func (i Inflight) GetBytes(key []byte) int64 {
// 	return i.est.GetBytes(key)
// }

// func (i Inflight) IncrString(key string) (int64, func()) {
// 	return i.IncrNString(key, 1)
// }

// func (i Inflight) IncrNString(key string, n int64) (int64, func()) {
// 	v := i.est.IncrNString(key, n)
// 	return v, func() { i.est.IncrNString(key, -n) }
// }

// func (i Inflight) IncrBytes(key []byte) (int64, func()) {
// 	return i.IncrNBytes(key, 1)
// }

// func (i Inflight) IncrNBytes(key []byte, n int64) (int64, func()) {
// 	v := i.est.IncrNBytes(key, n)
// 	return v, func() { i.est.IncrNBytes(key, -n) }
// }
