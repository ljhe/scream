package sessions

import (
	"github.com/gorilla/websocket"
	"github.com/ljhe/scream/3rd/logrus"
	"github.com/ljhe/scream/core/iface"
	"github.com/ljhe/scream/core/socket"
	"github.com/ljhe/scream/pbgo"
	"runtime/debug"
	"sync/atomic"
)

type WSSession struct {
	*Session
	conn *websocket.Conn
}

func (ws *WSSession) SetConn(c interface{}) {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	ws.conn = c.(*websocket.Conn)
}

func (ws *WSSession) Conn() interface{} {
	return ws.conn
}

func (ws *WSSession) CloseEvent(err error) {
	ws.Hooker.InEvent(&socket.RcvProcEvent{Sess: ws, Message: &pbgo.WSSessionClosedNtf{}, Err: err})
}

func (ws *WSSession) RunRcv() {
	defer func() {
		if err := recover(); err != nil {
			logrus.Errorf("tcpSession Stack---::%v\n %s\n", err, string(debug.Stack()))
			debug.PrintStack()
		}
	}()

	for {
		msg, err := ws.ReadMsg(ws)
		if err != nil {
			logrus.Errorf("RunRcv ReadMsg err:%v sessionId:%d", err, ws.GetId())
			// 做关闭处理 发送数据时已经无法发送
			atomic.StoreInt64(&ws.close, 1)
			select {
			case ws.sendQueue <- nil:
			default:
				logrus.Errorf("RunRcv sendQueue block len:%d sessionId:%d", len(ws.sendQueue), ws.GetId())
			}

			// 抛出关闭事件
			ws.CloseEvent(err)
			break
		}

		// 不同玩家之间并行处理
		ws.Hooker.InEvent(&socket.RcvProcEvent{Sess: ws, Message: msg})
	}
	ws.exitWg.Done()
}

func NewWSSession(conn *websocket.Conn, node iface.INetNode) *WSSession {
	ws := &WSSession{
		conn:    conn,
		Session: NewSession(node),
	}
	ws.ISessionExtension = ws
	return ws
}
