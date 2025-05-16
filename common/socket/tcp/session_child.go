package tcp

import (
	"github.com/ljhe/scream/common/socket"
	"github.com/ljhe/scream/plugins/logrus"
	"runtime/debug"
	"sync/atomic"
)

type SessionChild struct {
	sessionId uint64
	*session
	close    int64
	rcvQueue chan interface{}
}

func NewSessionChild(sessionId uint64, s *session) *SessionChild {
	return &SessionChild{
		sessionId: sessionId,
		session:   s,
		rcvQueue:  make(chan interface{}, 500),
	}
}

func (sc *SessionChild) Start(sessionId uint64) {
	atomic.StoreInt64(&sc.close, 0)
	go sc.RunRcv()
}

func (sc *SessionChild) Stop(sessionId uint64) {
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

func (sc *SessionChild) RunRcv() {
	defer func() {
		if err := recover(); err != nil {
			logrus.Log(logrus.LogsSystem).Errorf("session children Stack---::%v\n %s\n", err, string(debug.Stack()))
			debug.PrintStack()
		}
	}()

	for {
		for data := range sc.rcvQueue {
			if atomic.LoadInt64(&sc.close) == 1 {
				break
			}
			if data == nil {
				continue
			}
			sc.MsgRouter(&socket.RcvMsgEvent{Sess: sc.session, Message: data, Err: nil})
		}
	}
}
