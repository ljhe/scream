package tests

import (
	"context"
	"github.com/ljhe/scream/core/process"
	"github.com/ljhe/scream/msg"
	"github.com/ljhe/scream/tests/mock"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestReenter(t *testing.T) {

	p := process.BuildProcessWithOption(
		process.WithID("test-reenter-1"),
		process.WithLoader(loader),
		process.WithFactory(factory),
	)

	// build
	var err error
	_, err = p.System().Loader("mocka").WithID("mocka").Register(context.TODO())
	assert.Equal(t, err, nil)
	_, err = p.System().Loader("mockb").WithID("mockb").Register(context.TODO())
	assert.Equal(t, err, nil)

	p.Init()
	defer func() {
		wg := sync.WaitGroup{}
		p.System().Exit(&wg)
		wg.Wait()
	}()

	time.Sleep(time.Second)

	t.Run("Normal Case", func(t *testing.T) {
		mock.RecenterCalcValue = 0
		err := p.System().Call("mocka", "mocka", "reenter",
			msg.NewBuilder(context.TODO()).Build())
		assert.Nil(t, err)
		time.Sleep(time.Second)
		assert.Equal(t, int32(8), mock.RecenterCalcValue) // (2 + 2) * 2
	})

	t.Run("Timeout Case", func(t *testing.T) {

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		m := msg.NewBuilder(ctx).Build()

		p.System().Call("mocka", "mocka", "timeout", m)
		time.Sleep(time.Second * 4)
		assert.NotNil(t, m.Err)
	})

	t.Run("Timeout chain", func(t *testing.T) {
		mock.RecenterCalcValue = 0

		err := p.System().Call("mocka", "mocka", "chain", msg.NewBuilder(context.TODO()).Build())
		assert.Nil(t, err)

		assert.Nil(t, err)
		time.Sleep(time.Second)
		assert.Equal(t, int32(18), mock.RecenterCalcValue) // ((2 + 2) * 2) + 10
	})
}
