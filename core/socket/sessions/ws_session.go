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

func (ws *WSSession) RunRcv() {
	defer func() {
		// 打印堆栈信息
		if err := recover(); err != nil {
			logrus.Log(logrus.LogsSystem).Errorf("wsSession Stack---::%v\n %s\n", err, string(debug.Stack()))
			debug.PrintStack()
		}
	}()

	for {
		msg, err := ws.ReadMsg(ws)
		if err != nil {
			// 检测是否正常关闭
			if !websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				logrus.Log(logrus.LogsSystem).Errorf("RunRcv ReadMsg err:%v sessionId:%d \n", err, ws.GetId())
			}
			// 做关闭处理 发送数据时已经无法发送
			atomic.StoreInt64(&ws.close, 1)
			select {
			case ws.sendQueue <- nil:
			default:
				logrus.Log(logrus.LogsSystem).Errorf("RunRcv sendQueue block len:%d sessionId:%d \n", len(ws.sendQueue), ws.GetId())
			}

			// 抛出关闭事件
			ws.ProcEvent(&socket.RcvMsgEvent{Sess: ws, Message: &pbgo.WSSessionClosedNtf{}, Err: err})
			break
		}

		// 接收数据事件放到队列中(需要放到队列中，否则会有线程冲突)
		ws.ProcEvent(&socket.RcvMsgEvent{Sess: ws, Message: msg, Err: nil})
	}

	logrus.Log(logrus.LogsSystem).Infof("wsSession exit addr:%v", ws.conn.LocalAddr())
	ws.exitWg.Done()
}

func (ws *WSSession) TransmitChild(sessionId uint64, data interface{}) {

}

func (ws *WSSession) DelChild(sessionId uint64) {

}

func NewWSSession(conn *websocket.Conn, node iface.INetNode) *WSSession {
	ws := &WSSession{
		conn:    conn,
		Session: NewSession(node),
	}
	ws.ISessionExtension = ws
	node.(socket.Option).CopyOpt(&ws.sessionOpt)
	return ws
}
