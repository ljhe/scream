package tests

import (
	"context"
	"github.com/ljhe/scream/core/process"
	"github.com/ljhe/scream/def"
	"github.com/ljhe/scream/msg"
	"github.com/ljhe/scream/tests/mock"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// go test -benchmem -run=^$ -bench ^BenchmarkCall$ -v -benchtime=10s
//BenchmarkCall-20           64374            174503 ns/op           15518 B/op        241 allocs/op

// 屏蔽etcd相关操作后
//BenchmarkCall-20          165931             67673 ns/op            6569 B/op         82 allocs/op

func BenchmarkCall(b *testing.B) {
	p1 := process.BuildProcessWithOption(
		process.WithID("bench-call-1"),
		process.WithPort(8888),
		process.WithLoader(loader),
		process.WithFactory(factory),
	)

	p2 := process.BuildProcessWithOption(
		process.WithID("bench-call-2"),
		process.WithPort(7777),
		process.WithLoader(loader),
		process.WithFactory(factory),
	)

	// build
	p1.System().Loader("mocka").WithID("mocka").Register(context.TODO())
	p2.System().Loader("mockb").WithID("mockb").Register(context.TODO())

	p1.Init()
	p2.Init()
	defer func() {
		wg1 := sync.WaitGroup{}
		wg2 := sync.WaitGroup{}
		p1.System().Exit(&wg1)
		p2.System().Exit(&wg2)
		wg1.Wait()
		wg2.Wait()
	}()

	time.Sleep(time.Second)
	b.ResetTimer()

	atomic.StoreInt64(&mock.BechmarkCallReceivedMessageCount, 0)

	for i := 0; i < b.N; i++ {
		p1.System().Call(def.SymbolLocalFirst,
			"mocka",
			"call_benchmark",
			msg.NewBuilder(context.TODO()).WithReqBody([]byte{}).Build())
	}

	time.Sleep(time.Second)
	b.Logf("Total messages received: %d", atomic.LoadInt64(&mock.BechmarkCallReceivedMessageCount))
}
