package msg

import (
	"testing"
)

var mb MsgBase

func Benchmark_WithPool(b *testing.B) {
	MsgOptions.Pool = true

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		mb.ActualDataLen = i % 1024 // 控制最大值 防止slice无限扩大
		buf := mb.Container()
		mb.Release(buf)
	}
}

func Benchmark_WithOutPool(b *testing.B) {
	MsgOptions.Pool = false

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		mb.ActualDataLen = i % 1024
		mb.Container()
	}
}

// go test -bench=Benchmark_ -benchmem
// 方法名-并发线程数				   循环次数				 每次操作平均耗时    每次操作平均分配内存字节数  每次操作平均内存分配次数
// Benchmark_WithPool-20           30182679                33.71 ns/op           25 B/op          1 allocs/op
// Benchmark_WithOutPool-20        10550127               104.0 ns/op           540 B/op          0 allocs/op

// gen msg.proto
// .\pbgo\protoc --proto_path=./msg --go_out=./msg --go-grpc_out=./msg msg.proto
