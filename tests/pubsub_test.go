package tests

import (
	"context"
	"github.com/ljhe/scream/core/process"
	"github.com/ljhe/scream/tests/mock"
	"github.com/stretchr/testify/assert"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestPubsub(t *testing.T) {
	p := process.BuildProcessWithOption(
		process.WithID("test-pubsub-1"),
		process.WithLoader(loader),
		process.WithFactory(factory),
	)

	// build
	var err error
	_, err = p.System().Loader("mocka").WithID("mocka").Register(context.TODO())
	assert.Equal(t, err, nil)

	p.Init()
	defer func() {
		wg := sync.WaitGroup{}
		p.System().Exit(&wg)
		wg.Wait()
	}()

	t.Run("normal", func(t *testing.T) {
		time.Sleep(time.Second * 1)

		err = p.System().Pub("mocka", "offline_msg", []byte("offline msg"))
		assert.Equal(t, err, nil)

		time.Sleep(time.Second * 1)
	})
}

// go test -benchmem -run=^$ -bench ^BenchmarkPubsub$ -v -benchtime=10s
func BenchmarkPubsub(b *testing.B) {
	atomic.StoreInt64(&mock.ReceivedMessageCount, 0)

	p := process.BuildProcessWithOption(
		process.WithID("benchmark-pubsub-1"),
		process.WithLoader(loader),
		process.WithFactory(factory),
	)

	p.System().Loader("mocka").WithID("mocka").Register(context.TODO())

	p.Init()
	defer func() {
		wg := sync.WaitGroup{}
		p.System().Exit(&wg)
		wg.Wait()
	}()

	time.Sleep(time.Second)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		p.System().Pub("mocka", "offline_msg", []byte("offline msg"))
	}

	// 等待一小段时间确保消息都被处理
	time.Sleep(time.Second)
	b.Logf("Total messages received: %d", atomic.LoadInt64(&mock.ReceivedMessageCount))

}
