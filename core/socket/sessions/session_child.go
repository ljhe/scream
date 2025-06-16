package sessions

import (
	"github.com/ljhe/scream/3rd/logrus"
	"github.com/ljhe/scream/core/socket"
	"runtime/debug"
	"sync/atomic"
)

const SessionChildRcvQueueLen = 500

type SessionChild struct {
	sessionId uint64
	*Session
	socket.Processor
	close    int64
	rcvQueue chan interface{}
}

func NewSessionChild(sessionId uint64, s *Session) *SessionChild {
	return &SessionChild{
		sessionId: sessionId,
		Session:   s,
		Processor: socket.Processor{
			MsgFlow: new(socket.WSMsgFlow),
			Hooker:  new(socket.SessionChildHookEvent),
		},
		rcvQueue: make(chan interface{}, SessionChildRcvQueueLen),
	}
}

func (sc *SessionChild) Start() {
	atomic.StoreInt64(&sc.close, 0)
	go sc.RunRcv()
}

func (sc *SessionChild) Stop() {
	sc.Rcv(nil)
	atomic.StoreInt64(&sc.close, 1)
}

func (sc *SessionChild) Rcv(msg interface{}) {
	if atomic.LoadInt64(&sc.close) != 0 {
		return
	}
	select {
	case sc.rcvQueue <- msg:
	default:
	}
}

func (sc *SessionChild) GetSessionId() uint64 {
	return sc.sessionId
}

func (sc *SessionChild) RunRcv() {
	defer func() {
		if err := recover(); err != nil {
			logrus.Errorf("session children Stack---::%v\n %s\n", err, string(debug.Stack()))
			debug.PrintStack()
		}
	}()

	for data := range sc.rcvQueue {
		if atomic.LoadInt64(&sc.close) == 1 {
			break
		}
		if data == nil {
			break
		}
		sc.Processor.ProcEvent(&socket.RcvProcEvent{Sess: sc, Message: data, Err: nil})
	}
	logrus.Infof("session children close. sessionId:%d", sc.sessionId)
}
