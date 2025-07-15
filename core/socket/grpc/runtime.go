package grpc

import (
	"context"
	"fmt"
	"github.com/ljhe/scream/3rd/logrus"
	"github.com/ljhe/scream/core/iface"
	"github.com/ljhe/scream/lib/mpsc"
	"github.com/ljhe/scream/pbgo"
)

type Runtime struct {
	q      *mpsc.Queue
	chains map[string]iface.IChain
}

func (r *Runtime) Init(ctx context.Context) {
	r.q = mpsc.New()
	r.chains = make(map[string]iface.IChain)
	go r.update()
}

func (r *Runtime) OnEvent(ev string, chainFunc func() iface.IChain) error {
	if _, exists := r.chains[ev]; exists {
		return fmt.Errorf("actor: repeat register event %v", ev)
	}
	r.chains[ev] = chainFunc()
	return nil
}

func (r *Runtime) Received(mw interface{}) error {
	r.q.Push(mw)
	return nil
}

func (r *Runtime) update() {
	for {
		select {
		case <-r.q.C:
			msgInterface := r.q.Pop()
			msg, ok := msgInterface.(*pbgo.RouteReqs)
			if !ok {
				continue
			}
			if chain, ok := r.chains[msg.Msg.Header.Event]; ok {
				err := chain.Execute()
				if err != nil {
					logrus.Errorf("event: [%s] execute chain err %v", msg.Msg.Header.Event, err)
				}
			} else {
				logrus.Errorf("grpc message event: [%v], not found in chain", msg.Msg.Header.Event)
			}
		}
	}
}
