package sessions

import (
	"github.com/ljhe/scream/core/iface"
	"github.com/ljhe/scream/core/socket"
	"net"
	"sync/atomic"
	"time"
)

const (
	HeartBeatNewTimer = 15 * time.Second
	HeartBeatReset    = 5 * time.Second
)

type TCPSession struct {
	*Session
	conn net.Conn
}

func (ts *TCPSession) SetConn(c interface{}) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.conn = c.(net.Conn)
}

func (ts *TCPSession) Conn() interface{} {
	return ts.conn
}

func (ts *TCPSession) TransmitChild(sessionId uint64, data interface{}) {
	var sc *SessionChild
	if val, ok := ts.children.Load(sessionId); !ok {
		sc = NewSessionChild(sessionId, ts.Session)
		sc.Start()
		ts.children.Store(sessionId, sc)
	} else {
		sc = val.(*SessionChild)
	}
	sc.Rcv(data)
}

func (ts *TCPSession) DelChild(sessionId uint64) {
	if val, ok := ts.children.Load(sessionId); ok {
		val.(*SessionChild).Stop()
		ts.children.Delete(sessionId)
	}
}

// HeartBeat 服务器之间的心跳检测
func (ts *TCPSession) HeartBeat(msg interface{}) {
	if atomic.LoadInt64(&ts.close) != 0 {
		return
	}

	go func() {
		delayTimer := time.NewTimer(HeartBeatNewTimer)
		for {
			delayTimer.Reset(HeartBeatReset)
			select {
			case <-delayTimer.C:
				if atomic.LoadInt64(&ts.close) != 0 {
					break
				}
				ts.Send(msg)
			}
		}
	}()
}

func (ts *TCPSession) CloseEvent(err error) {
	ts.ProcEvent(&socket.RcvProcEvent{Sess: ts, Message: &socket.SessionClosed{}, Err: err})
}

func NewTcpSession(c net.Conn, node iface.INetNode) *TCPSession {
	ts := &TCPSession{
		conn:    c,
		Session: NewSession(node),
	}
	ts.ISessionExtension = ts
	return ts
}
