package log

import (
	"testing"
)

func init() {
	NewDefaultLogger()
}

// benchmark text not written to file
// 56005             28258 ns/op             209 B/op          5 allocs/op

// benchmark text written to file
// 205292             6665 ns/op             594 B/op         12 allocs/op

// benchmark json not written to file
// 62550             20639 ns/op             208 B/op          2 allocs/op

// benchmark json written to file
// 156963             6874 ns/op             529 B/op          9 allocs/op

func Benchmark_ZapLog(b *testing.B) {
	for i := 0; i < b.N; i++ {
		InfoF("this is test: %d", i)
		//InfoJ("this is test: %d", i)
	}
}
