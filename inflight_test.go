package limits

// import (
// 	"strconv"
// 	"sync"
// 	"testing"
// )

// func TestInflight(t *testing.T) {
// 	inf := NewInflight()

// 	for i := 0; i < 10; i++ {
// 		func() {
// 			n, decr := inf.IncrString(strconv.Itoa(i + 1))
// 			defer decr()

// 			if n != 1 {
// 				t.Fatalf("unexpected value: %d", n)
// 			}
// 		}()
// 	}

// 	for i := 0; i < 10; i++ {
// 		n := inf.GetString(strconv.Itoa(i + 1))
// 		if n != 0 {
// 			t.Fatalf("unexpected value: %d", n)
// 		}
// 	}

// 	var fns []func()
// 	for i := 0; i < 10; i++ {
// 		n, decr := inf.IncrString("key")
// 		if n != int64(i)+1 {
// 			t.Fatalf("unexpected value: %d", n)
// 		}
// 		fns = append(fns, decr)
// 	}
// 	for _, fn := range fns {
// 		fn()
// 	}
// 	n := inf.GetString("key")
// 	if n != 0 {
// 		t.Fatalf("unexpected value: %d", n)
// 	}
// }

// func TestInflightConcurrency(t *testing.T) {
// 	inf := NewInflight()

// 	var wg sync.WaitGroup
// 	for i := 0; i < 10; i++ {
// 		wg.Add(1)
// 		go func() {
// 			defer wg.Done()
// 			for j := 0; j < 100; j++ {
// 				func() {
// 					_, decr := inf.IncrString(strconv.Itoa(j + 1))
// 					defer decr()
// 				}()
// 			}
// 		}()
// 	}
// 	wg.Wait()
// }
