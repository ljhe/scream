package tests

import (
	"context"
	"github.com/ljhe/scream/core/process"
	"github.com/ljhe/scream/msg"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSend(t *testing.T) {
	p := process.BuildProcessWithOption(
		process.WithID("test-send-1"),
		process.WithLoader(loader),
		process.WithFactory(factory),
	)

	// build
	var err error
	_, err = p.System().Loader("mockb").WithID("mockb").Register(context.TODO())
	assert.Equal(t, err, nil)

	p.Init()
	defer func() {
		wg := sync.WaitGroup{}
		p.System().Exit(&wg)
		wg.Wait()
	}()

	t.Run("normal", func(t *testing.T) {
		m := msg.NewBuilder(context.TODO()).Build()
		timenow := time.Now()
		err := p.System().Send("mockb", "mockb", "timeout", m)

		assert.Equal(t, true, time.Since(timenow) < time.Second) // 添加时间差检查
		assert.Equal(t, err, nil)
	})
}
