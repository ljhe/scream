package tests

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ljhe/scream/core"
	"github.com/ljhe/scream/core/node"
	"github.com/ljhe/scream/def"
	"github.com/ljhe/scream/router/msg"
	"github.com/ljhe/scream/tests/mock"
)

// go test -benchmem -run=^$ -bench ^BenchmarkCall$ github.com/ljhe/scream/tests -v -benchtime=10s
func BenchmarkCall(b *testing.B) {
	nod1 := node.BuildProcessWithOption(
		core.NodeWithID("bench-call-1"),
		core.NodeWithPort(8888),
		core.NodeWithLoader(loader),
		core.NodeWithFactory(factory),
	)

	nod2 := node.BuildProcessWithOption(
		core.NodeWithID("bench-call-2"),
		core.NodeWithPort(7777),
		core.NodeWithLoader(loader),
		core.NodeWithFactory(factory),
	)

	// build
	nod1.System().Loader("mocka").WithID("mocka").Register(context.TODO())
	nod2.System().Loader("mockb").WithID("mockb").Register(context.TODO())

	nod1.Init()
	nod2.Init()
	defer func() {
		wg1 := sync.WaitGroup{}
		wg2 := sync.WaitGroup{}
		nod1.System().Exit(&wg1)
		nod2.System().Exit(&wg2)
		wg1.Wait()
		wg2.Wait()
	}()

	time.Sleep(time.Second)
	b.ResetTimer()

	atomic.StoreInt64(&mock.BechmarkCallReceivedMessageCount, 0)

	for i := 0; i < b.N; i++ {
		nod1.System().Call(def.SymbolLocalFirst,
			"mocka",
			"call_benchmark",
			msg.NewBuilder(context.TODO()).WithReqBody([]byte{}).Build())
	}

	time.Sleep(time.Second)
	b.Logf("Total messages received: %d", atomic.LoadInt64(&mock.BechmarkCallReceivedMessageCount))
}
