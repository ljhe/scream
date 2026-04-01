package buffer

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestBufferWrites(t *testing.T) {
	buf := &Buffer{bs: make([]byte, defaultSize, defaultSize)}
	fmt.Printf("%p\n", buf.bs)
	buf.Reset(8)
	fmt.Printf("%p\n", buf.bs)
	buf.Reset(2049)
	fmt.Printf("%p\n", buf.bs)
}

func BenchmarkBuffers(b *testing.B) {
	// Because we use the strconv.AppendFoo functions so liberally, we can't
	// use the standard library's bytes.Buffer anyways (without incurring a
	// bunch of extra allocations). Nevertheless, let's make sure that we're
	// not losing any precious nanoseconds.
	str := strings.Repeat("a", 1024)
	strBytes := []byte(str)
	slice := make([]byte, 1024)
	buf := bytes.NewBuffer(slice)
	custom := NewPool().GetWithSize(1024)
	b.Run("ByteSlice", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			slice = append(slice, str...)
			slice = slice[:0]
		}
	})
	b.Run("BytesBuffer", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			buf.WriteString(str)
			buf.Reset()
		}
	})
	b.Run("CustomBuffer", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			custom.Write(strBytes)
			custom.Reset(1024)
		}
	})
}
