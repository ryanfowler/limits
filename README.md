# limits

Lock-free, space efficient, probabilistic data structures for counting things.

## Usage

### `Estimator`

An `Estimator` is an implementation of the [count-min sketch](https://en.wikipedia.org/wiki/Count%E2%80%93min_sketch)
data structure. It can be used to estimate the count of occurrences for
different keys.

```go
est := NewEstimator[string]()

n := est.Incr("key")
println(n) // Prints "1"

n = est.Get("key")
println(n) // Prints "1"
```

### `Rate`

`Rate` allows for the estimation of occurrences for different keys in a given
interval. Internally, it uses two `Estimator`s to keep track of the current
interval, as well as one full interval in the past.

Observing an event will increment the count for the provided key, and return
the estimated number of occurences in the current interval.

```go
r := NewRate[string](time.Second)

n := r.Observe("key")
println(n) // Prints "1"
```

You may also obtain the estimated number of occurences in the last full interval
by using the `Get` method.

```go
r := NewRate[string](time.Second)

n := r.Observe("key"")
println(n) // Prints "1"
n = r.Get("key")
println(n) // Prints "0"

time.Sleep(time.Second)
n = r.Get("key")
println(n) // Prints "1"
```

A full rate limiting implementation will likely want to check the value of the
current interval, as well as the last "full" interval. This is because the
current interval may have just started, and thus the value may not be accurate
as an estimation of the actual number of occurences per interval period.

```go
// Allow 10 requests per second
const rps = 10

var limiter = NewRate[string](time.Second)

func AllowRequest(key string) bool {
	// Increment count, and check the total in the current interval.
	if n := limiter.Observe("key"); n > rps {
		return false
	}

	// Check the count in the last full interval.
	if n := limiter.Get("key"); n > rps {
		return false
	}

	return true
}
```

## License

Apache-2.0 license
