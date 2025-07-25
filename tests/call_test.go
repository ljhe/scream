package tests

import (
	"context"
	"github.com/ljhe/scream/core/process"
	"github.com/ljhe/scream/msg"
	"github.com/ljhe/scream/tests/mock"
	"github.com/stretchr/testify/assert"
	"math/rand/v2"
	"sync"
	"testing"
)

func TestCall(t *testing.T) {
	p := process.BuildProcessWithOption(
		process.WithLoader(loader),
		process.WithLoader(loader),
		process.WithFactory(factory),
	)

	_, err := p.System().Loader("mocka").WithID("mocka").WithType("mocka").Register(context.TODO())
	assert.Equal(t, err, nil)
	_, err = p.System().Loader("mockb").WithID("mockb").WithType("mockb").Register(context.TODO())
	assert.Equal(t, err, nil)
	_, err = p.System().Loader("mockc").WithID("mockc").WithType("mockc").Register(context.TODO())
	assert.Equal(t, err, nil)

	p.Init()
	defer func() {
		wg := sync.WaitGroup{}
		p.System().Exit(&wg)
		wg.Wait()
	}()

	t.Run("normal", func(t *testing.T) {
		m := msg.NewBuilder(context.TODO()).Build()
		err := p.System().Call("mockc", "mockc", "ping", m)
		assert.Equal(t, err, nil)

		resval := msg.GetResCustomField[string](m, "pong")
		assert.Equal(t, resval, "pong")
	})
}

func TestCallBlock(t *testing.T) {
	p := process.BuildProcessWithOption(
		process.WithID("test-call-block"),
		process.WithLoader(loader),
		process.WithFactory(factory),
	)

	// build
	var err error
	_, err = p.System().Loader("mocka").WithID("mocka").Register(context.TODO())
	assert.Equal(t, err, nil)
	_, err = p.System().Loader("mockb").WithID("mockb").Register(context.TODO())
	assert.Equal(t, err, nil)
	_, err = p.System().Loader("mockc").WithID("mockc").Register(context.TODO())
	assert.Equal(t, err, nil)

	p.Init()
	defer func() {
		wg := sync.WaitGroup{}
		p.System().Exit(&wg)
		wg.Wait()
	}()

	// a (+1 -> b (+1 -> c (+1
	t.Run("normal", func(t *testing.T) {
		m := msg.NewBuilder(context.TODO())

		r := rand.IntN(10)
		m.WithReqCustomFields(msg.Attr{Key: "randvalue", Value: r})
		err := p.System().Call("mocka", "mocka", "test_block", m.Build())
		assert.Equal(t, err, nil)

		resval := msg.GetResCustomField[int](m.Build(), "randvalue")
		assert.Equal(t, resval, r+3)
	})
}

func TestTCCSucc(t *testing.T) {
	p := process.BuildProcessWithOption(
		process.WithID("test-tcc-1"),
		process.WithLoader(loader),
		process.WithFactory(factory),
	)

	// build
	var err error
	_, err = p.System().Loader("mocka").WithID("mocka").Register(context.TODO())
	assert.Equal(t, err, nil)
	_, err = p.System().Loader("mockb").WithID("mockb").Register(context.TODO())
	assert.Equal(t, err, nil)
	_, err = p.System().Loader("mockc").WithID("mockc").Register(context.TODO())
	assert.Equal(t, err, nil)

	p.Init()
	defer func() {
		wg := sync.WaitGroup{}
		p.System().Exit(&wg)
		wg.Wait()
	}()

	err = p.System().Call("mocka", "mocka", "tcc_succ", msg.NewBuilder(context.TODO()).Build())
	assert.Nil(t, err)

	assert.Equal(t, mock.MockBTccValue, 111)
	assert.Equal(t, mock.MockCTccValue, 222)
}
