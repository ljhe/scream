package node

import (
	"context"
	"fmt"
	"github.com/ljhe/scream/3rd/logrus"
	"github.com/ljhe/scream/core/iface"
	"github.com/ljhe/scream/core/system"
	"github.com/ljhe/scream/lib/mpsc"
	"github.com/ljhe/scream/lib/pubsub"
	"github.com/ljhe/scream/router"
	"reflect"
	"runtime/debug"
	"sync/atomic"
	"time"
)

type RecoveryFunc func(interface{})

type Node struct {
	Id           string
	Ty           string
	Sys          iface.ISystem
	q            *mpsc.Queue
	reenterQueue *mpsc.Queue
	closed       int32
	closeCh      chan struct{}
	shutdownCh   chan struct{}
	chains       map[string]iface.IChain
	recovery     RecoveryFunc

	timers    map[iface.ITimer]struct{}
	timerChan chan iface.ITimer
	//timerWg   sync.WaitGroup // 用于等待所有 timer goroutine 退出

	actorCtx *nodeContext
}

func (n *Node) ID() string {
	return n.Id
}

func (n *Node) Type() string {
	return n.Ty
}

func (n *Node) Init(ctx context.Context) {
	n.q = mpsc.New()
	n.reenterQueue = mpsc.New()
	atomic.StoreInt32(&n.closed, 0) // 初始化closed状态为0（未关闭）
	n.closeCh = make(chan struct{})
	n.shutdownCh = make(chan struct{})
	n.chains = make(map[string]iface.IChain)
	n.recovery = defaultRecovery
	n.actorCtx = &nodeContext{
		ctx: ctx,
	}

	n.actorCtx.ctx = context.WithValue(n.actorCtx.ctx, systemKey{}, n.Sys)
	n.actorCtx.ctx = context.WithValue(n.actorCtx.ctx, actorKey{}, n)

	n.timers = make(map[iface.ITimer]struct{})
	n.timerChan = make(chan iface.ITimer, 1024)

	go n.update()
}

func defaultRecovery(r interface{}) {
	logrus.Errorf("node Recovered from panic: %v\nStack trace:\n%s\n", r, debug.Stack())
}

func (n *Node) Context() iface.INodeContext {
	return n.actorCtx
}

func (n *Node) OnEvent(ev string, chainFunc func(iface.INodeContext) iface.IChain) error {
	if _, exists := n.chains[ev]; exists {
		return fmt.Errorf("actor: repeat register event %v", ev)
	}
	n.chains[ev] = chainFunc(n.actorCtx)
	return nil
}

// OnTimer register timer
//
//	dueTime: Delay time before starting the timer (in milliseconds). If 0, starts immediately
//	interval: Time interval between executions (in milliseconds). If 0, executes only once
//	f: Callback function
//	args: Arguments for the callback function
func (n *Node) OnTimer(dueTime int64, interval int64, f func(interface{}) error, args interface{}) iface.ITimer {
	info := NewTimerInfo(
		time.Duration(dueTime)*time.Millisecond,
		time.Duration(interval)*time.Millisecond,
		f, args)

	n.timers[info] = struct{}{}
	//a.timerWg.Add(1)

	go func() {
		//defer a.timerWg.Done()

		// 如果 dueTime 大于 0，使用 dueTime 进行第一次触发
		if info.dueTime > 0 {
			<-time.After(info.dueTime)
			n.timerChan <- info
		}

		info.ticker = time.NewTicker(info.interval)

		for {
			select {
			case <-info.ticker.C:
				n.timerChan <- info
			case <-n.shutdownCh:
				return
			}
		}
	}()

	return info
}

func (n *Node) CancelTimer(t iface.ITimer) {
	if t == nil {
		return
	}

	logrus.Infof("[braid.timer] %v timer cancel", n.Id)

	t.Stop()
	delete(n.timers, t)
}

// Sub subscribes to a message
//
//	If this is the first subscription to this topic, opts will take effect (you can set some options for the topic, such as ttl)
//	topic: A subject that contains a group of channels (e.g., if topic = offline messages, channel = actorId, then each actor can get its own offline messages in this topic)
//	channel: Represents different categories within a topic
//	callback: Callback function for successful subscription
func (n *Node) Sub(topic string, channel string, callback func(ctx iface.INodeContext) iface.IChain, opts ...pubsub.TopicOption) error {

	ch, err := n.Sys.Sub(topic, channel, opts...)
	if err != nil {
		return fmt.Errorf("failed to subscribe to topic %s: %w", topic, err)
	}

	ch.Arrived(n.q)

	n.OnEvent(channel, callback)

	return nil
}

func (n *Node) Call(idOrSymbol, actorType, event string, mw *router.Wrapper) error {

	if mw.Req.Header.OrgActorID == "" { // Only record the original sender
		mw.Req.Header.OrgActorID = n.Id
		mw.Req.Header.OrgActorType = n.Ty
	}

	// Updated to the latest value on each call
	mw.Req.Header.PrevActorType = n.Ty

	return n.Sys.Call(idOrSymbol, actorType, event, mw)
}

func (n *Node) Received(mw *router.Wrapper) error {
	if mw.Req.Header.OrgActorID != "" {
		if mw.Req.Header.OrgActorID == n.Id {
			return system.ErrSelfCall
		}
	}

	if atomic.LoadInt32(&n.closed) != 0 {
		// Actor已关闭，不处理消息，也不增加计数器
		logrus.Warnf("actor %v is closed, message %v will be ignored", n.Id, mw.Req.Header.Event)
		return fmt.Errorf("actor %v is closed", n.Id)
	}

	mw.GetWg().Add(1)
	n.q.Push(mw)
	return nil
}

func (n *Node) ReenterCall(idOrSymbol, actorType, event string, rmw *router.Wrapper) iface.IFuture {
	if rmw.Req.Header.OrgActorID == "" {
		rmw.Req.Header.OrgActorID = n.Id
		rmw.Req.Header.OrgActorType = n.Ty
	}
	rmw.Req.Header.PrevActorType = n.Ty

	reenterFuture := NewFuture()
	callFuture := NewFuture()

	deadline, ok := rmw.Ctx.Deadline()
	var timeout time.Duration
	if ok {
		timeout = time.Until(deadline)
	} else {
		timeout = 30 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	go func() {
		select {
		case <-rmw.Ctx.Done():
			logrus.Infof("ReenterCall Context canceled: %v", rmw.Ctx.Err())
			cancel()

			errWrapper := &router.Wrapper{
				Ctx: rmw.Ctx,
				Err: rmw.Ctx.Err(),
			}

			reenterFuture.Complete(errWrapper)
		case <-callFuture.done:
		}
	}()

	go func() {
		defer cancel()
		logrus.Infof("ReenterCall Starting call to %s.%s", actorType, event)

		swappedWrapper := router.Swap(rmw)
		//swappedWrapper.Ctx = ctx

		err := n.Sys.Call(idOrSymbol, actorType, event, swappedWrapper)
		if err != nil {
			callFuture.Complete(&router.Wrapper{
				Ctx: ctx,
				Err: err,
			})
			return
		}

		callFuture.Complete(swappedWrapper)
	}()

	// 设置回调，将处理放入重入队列
	callFuture.Then(func(ret *router.Wrapper) {

		reenterMsg := &reenterMessage{
			action: func(mw *router.Wrapper) error {

				defer func() {
					if r := recover(); r != nil {
						logrus.Errorf("panic in ReenterCall: %v", r)
						rmw.Err = fmt.Errorf("panic in ReenterCall: %v", r)
						reenterFuture.Complete(rmw)
					}
				}()

				if mw.Err != nil {
					rmw.Err = mw.Err
					reenterFuture.Complete(rmw)
					return mw.Err
				}

				rmw.Res = mw.Res
				reenterFuture.Complete(rmw)
				return nil
			},
			msg: ret,
		}

		n.reenterQueue.Push(reenterMsg)
	})

	return reenterFuture
}

func (n *Node) update() {
	checkClose := func() {
		timeout := time.After(10 * time.Second)
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()

		for !n.q.Empty() || !n.reenterQueue.Empty() {
			select {
			case <-timeout:
				logrus.Errorf("actor %s force close due to timeout waiting for queue to empty remaining %v", n.Id, n.q.Count())
				goto ForceClose
			case <-ticker.C:
				continue
			}
		}

	ForceClose:
		if atomic.CompareAndSwapInt32(&n.closed, 1, 2) {
			logrus.Infof("actor %s closing channel", n.Id)
			close(n.closeCh)
		}
	}

	for {
		select {
		case timerInfo := <-n.timerChan:
			if atomic.LoadInt32(&n.closed) != 0 {
				continue
			}
			func() {
				defer func() {
					if r := recover(); r != nil {
						n.recovery(r)
					}
				}()

				if err := timerInfo.Execute(); err != nil {
					logrus.Errorf("actor %v timer callback error: %v", n.Id, err)
				}
			}()
		case <-n.q.C:
			msgInterface := n.q.Pop()

			mw, ok := msgInterface.(*router.Wrapper)
			if !ok {
				logrus.Errorf("actor %v received non-Message type %v", n.Id, reflect.TypeOf(msgInterface))
				continue
			}

			func() {
				defer func() {
					if r := recover(); r != nil {
						n.recovery(r)
					}

					mw.GetWg().Done()
				}()

				if chain, ok := n.chains[mw.Req.Header.Event]; ok {
					err := chain.Execute(mw)
					if err != nil {
						logrus.Errorf("actor %v event %v execute err %v", n.Id, mw.Req.Header.Event, err)
					}
				} else {
					logrus.Errorf("actor %v No handlers for message type: %s", n.Id, mw.Req.Header.Event)
				}
			}()

		case <-n.reenterQueue.C:
			reenterMsgInterface := n.reenterQueue.Pop()
			if reenterMsg, ok := reenterMsgInterface.(*reenterMessage); ok {
				reenterMsg.action(reenterMsg.msg.(*router.Wrapper))
			}

		case <-n.shutdownCh:
			if atomic.CompareAndSwapInt32(&n.closed, 0, 1) {
				logrus.Infof("actor %s exiting check close %v", n.Id, atomic.LoadInt32(&n.closed))
				go checkClose()
			}
		case <-n.closeCh:
			logrus.Infof("[braid.actor] %s exiting closed", n.Id)
			return
		}
	}
}

func (n *Node) Exit() {
	logrus.Infof("node %s exiting", n.Id)
}
