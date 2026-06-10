package buffer

import (
	"bytes"
	"strings"
	"sync"
	"testing"
)

func TestBuffers(t *testing.T) {
	const dummyData = "dummy data"
	p := NewPool()

	var wg sync.WaitGroup
	for g := 0; g < 10; g++ {
		wg.Add(1)
		go func() {
			for i := 0; i < 100; i++ {
				buf := p.GetWithSize(1024)

				buf.Free()
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func BenchmarkPool_Get(b *testing.B) {
	// Because we use the strconv.AppendFoo functions so liberally, we can't
	// use the standard library's bytes.Buffer anyways (without incurring a
	// bunch of extra allocations). Nevertheless, let's make sure that we're
	// not losing any precious nanoseconds.
	str := strings.Repeat("a", 1024)
	strBytes := []byte(str)
	pool := NewPool()
	b.Run("ByteSlice", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			slice := make([]byte, 1024)
			slice = append(slice, str...)
		}
	})
	b.Run("BytesBuffer", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			slice := make([]byte, 1024)
			buf := bytes.NewBuffer(slice)
			buf.WriteString(str)
		}
	})
	b.Run("CustomBuffer", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			custom := pool.GetWithSize(1024)
			custom.Write(strBytes)
			custom.Free()
		}
	})
	b.Run("async ByteSlice", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				slice := make([]byte, 1024)
				slice = append(slice, str...)
			}
		})
	})
	b.Run("async CustomBuffer", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				custom := pool.GetWithSize(1024)
				custom.Write(strBytes)
				custom.Free()
			}
		})
	})

	/* # go test -run=xxx -bench=. -benchmem
	goos: windows
	goarch: amd64
	pkg: component/buffer
	BenchmarkPool_Get/ByteSlice-8            1644387               679 ns/op            2688 B/op          1 allocs/op
	BenchmarkPool_Get/BytesBuffer-8          1566680               692 ns/op            3072 B/op          1 allocs/op
	BenchmarkPool_Get/CustomBuffer-8        33422272                40.4 ns/op             0 B/op          0 allocs/op
	BenchmarkPool_Get/async_ByteSlice-8      1725154               677 ns/op            2688 B/op          1 allocs/op
	BenchmarkPool_Get/async_CustomBuffer-8  98485353                11.8 ns/op             0 B/op          0 allocs/op
	*/
}
