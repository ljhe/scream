package sessions

import (
	"github.com/gorilla/websocket"
	"github.com/ljhe/scream/core/iface"
	"github.com/ljhe/scream/core/socket"
	"github.com/ljhe/scream/pbgo"
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
	ws.ProcEvent(&socket.RcvProcEvent{Sess: ws, Message: &pbgo.WSSessionClosedNtf{}, Err: err})
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
