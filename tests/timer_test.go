package tests

import (
	"context"
	"fmt"
	"github.com/ljhe/scream/core/iface"
	"github.com/ljhe/scream/core/node"
	"github.com/ljhe/scream/core/process"
	"github.com/ljhe/scream/tests/mock"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockTimerActor struct {
	*node.Node
}

func newMockTimerActor(p iface.INodeBuilder) iface.INode {
	return &mockTimerActor{
		Node: &node.Node{Id: p.GetID(), Ty: p.GetType(), Sys: p.GetSystem()},
	}
}

var tick1 int32
var tick2 int32
var tick3 int32
var tick4 int32
var tick5 int32

func (ta *mockTimerActor) Init(ctx context.Context) {
	ta.Node.Init(ctx)

	ta.OnTimer(0, 1000, func(i interface{}) error {
		atomic.AddInt32(&tick1, 1)
		return nil
	}, nil)

	ta.OnTimer(500, 500, func(i interface{}) error {
		atomic.AddInt32(&tick2, 1)
		return nil
	}, nil)

	ta.OnTimer(0, 100, func(i interface{}) error {
		atomic.AddInt32(&tick3, 1)
		return nil
	}, nil)

	ta.OnTimer(1000, 0, func(i interface{}) error {
		atomic.AddInt32(&tick4, 1)
		return nil
	}, nil)

	var t iface.ITimer
	t = ta.OnTimer(0, 200, func(i interface{}) error {
		atomic.AddInt32(&tick5, 1)

		if atomic.LoadInt32(&tick5) == 5 {
			ta.CancelTimer(t)
		}

		return nil
	}, nil)
}

func TestActorTimer1(t *testing.T) {

	factory := mock.BuildNodeFactory()
	factory.Constructors["MockTimerActor"] = &iface.NodeConstructor{
		ID:                  "MockTimerActor",
		Name:                "MockTimerActor",
		Weight:              20,
		Constructor:         newMockTimerActor,
		NodeUnique:          false,
		GlobalQuantityLimit: 1,
		Dynamic:             false,
		Options:             make(map[string]string),
	}
	loader := node.BuildDefaultNodeLoader(factory)

	p := process.BuildProcessWithOption(
		process.WithID("test-timer-1"),
		process.WithLoader(loader),
		process.WithFactory(factory),
	)

	p.Init()

	t.Run("tick1", func(t *testing.T) {
		time.Sleep(time.Second * 5)
		tickcnt := atomic.LoadInt32(&tick1)
		assert.True(t, tickcnt >= int32(4) && tickcnt <= int32(6))
	})
}

func TestActorTimer2(t *testing.T) {

	factory := mock.BuildNodeFactory()
	factory.Constructors["MockTimerActor"] = &iface.NodeConstructor{
		ID:                  "MockTimerActor",
		Name:                "MockTimerActor",
		Weight:              20,
		Constructor:         newMockTimerActor,
		NodeUnique:          false,
		GlobalQuantityLimit: 1,
		Dynamic:             false,
		Options:             make(map[string]string),
	}
	loader := node.BuildDefaultNodeLoader(factory)

	p := process.BuildProcessWithOption(
		process.WithID("test-timer-2"),
		process.WithLoader(loader),
		process.WithFactory(factory),
	)

	p.Init()

	t.Run("tick2", func(t *testing.T) {
		time.Sleep(time.Second * 5)
		tickcnt := atomic.LoadInt32(&tick2)
		targetcnt := int32(5*(1000/500) - 1)
		assert.True(t, tickcnt >= int32(targetcnt-1) && tickcnt <= int32(targetcnt+1))
	})
}

func TestActorTimer3(t *testing.T) {

	factory := mock.BuildNodeFactory()
	factory.Constructors["MockTimerActor"] = &iface.NodeConstructor{
		ID:                  "MockTimerActor",
		Name:                "MockTimerActor",
		Weight:              20,
		Constructor:         newMockTimerActor,
		NodeUnique:          false,
		GlobalQuantityLimit: 1,
		Dynamic:             false,
		Options:             make(map[string]string),
	}
	loader := node.BuildDefaultNodeLoader(factory)

	p := process.BuildProcessWithOption(
		process.WithID("test-timer-3"),
		process.WithLoader(loader),
		process.WithFactory(factory),
	)

	p.Init()

	t.Run("tick3", func(t *testing.T) {
		time.Sleep(time.Second * 5)
		tickcnt := atomic.LoadInt32(&tick3)
		targetcnt := int32(5 * (1000 / 100))
		fmt.Println(tickcnt, targetcnt)
		assert.True(t, tickcnt >= int32(targetcnt-1) && tickcnt <= int32(targetcnt+1))
	})
}

func TestActorTimer4(t *testing.T) {

	factory := mock.BuildNodeFactory()
	factory.Constructors["MockTimerActor"] = &iface.NodeConstructor{
		ID:                  "MockTimerActor",
		Name:                "MockTimerActor",
		Weight:              20,
		Constructor:         newMockTimerActor,
		NodeUnique:          false,
		GlobalQuantityLimit: 1,
		Dynamic:             false,
		Options:             make(map[string]string),
	}
	loader := node.BuildDefaultNodeLoader(factory)

	p := process.BuildProcessWithOption(
		process.WithID("test-timer-4"),
		process.WithLoader(loader),
		process.WithFactory(factory),
	)

	p.Init()

	t.Run("tick4", func(t *testing.T) {
		time.Sleep(time.Second * 3)
		assert.Equal(t, atomic.LoadInt32(&tick4), int32(1))
	})
}

func TestActorTimer5(t *testing.T) {

	factory := mock.BuildNodeFactory()
	factory.Constructors["MockTimerActor"] = &iface.NodeConstructor{
		ID:                  "MockTimerActor",
		Name:                "MockTimerActor",
		Weight:              20,
		Constructor:         newMockTimerActor,
		NodeUnique:          false,
		GlobalQuantityLimit: 1,
		Dynamic:             false,
		Options:             make(map[string]string),
	}
	loader := node.BuildDefaultNodeLoader(factory)

	p := process.BuildProcessWithOption(
		process.WithID("test-timer-5"),
		process.WithLoader(loader),
		process.WithFactory(factory),
	)

	p.Init()

	t.Run("tick5", func(t *testing.T) {
		time.Sleep(time.Second * 3)
		assert.Equal(t, atomic.LoadInt32(&tick5), int32(5))
	})
}
