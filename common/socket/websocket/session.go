package websocket

import (
	"common/iface"
	"common/socket"
	"github.com/gorilla/websocket"
	"sync"
	"sync/atomic"
)

var sendQueueMaxLen = 2000

type wsSession struct {
	sync.Mutex
	*socket.NetProcessorRPC // 事件处理相关 procrpc.go
	socket.NetContextSet    // 记录session绑定信息 nodeproperty.go
	node                    iface.INetNode
	conn                    *websocket.Conn

	exitWg      sync.WaitGroup
	endCallback func()
	closeInt    int64

	sessionOpt socket.NetTCPSocketOption

	sendQueue       chan interface{}
	sendQueueMaxLen int
}

func (s *wsSession) SetConn(c *websocket.Conn) {
	s.Lock()
	defer s.Unlock()
	s.conn = c
}

func (s *wsSession) GetConn() *websocket.Conn {
	s.Lock()
	defer s.Unlock()
	return s.conn
}

func (s *wsSession) Raw() interface{} {
	return s.GetConn()
}

func (s *wsSession) Node() iface.INetNode {
	return s.node
}

func (s *wsSession) Send(msg interface{}) {
	//TODO implement me
	panic("implement me")
}

func (s *wsSession) Close() {
	//TODO implement me
	panic("implement me")
}

func (s *wsSession) SetId(id uint64) {
	//TODO implement me
	panic("implement me")
}

func (s *wsSession) GetId() uint64 {
	//TODO implement me
	panic("implement me")
}

func (s *wsSession) HeartBeat(msg interface{}) {
	//TODO implement me
	panic("implement me")
}

func (s *wsSession) IncRcvPingNum(inc int) {
	//TODO implement me
	panic("implement me")
}

func (s *wsSession) RcvPingNum() int {
	//TODO implement me
	panic("implement me")
}

func (s *wsSession) setConn(c *websocket.Conn) {
	s.Lock()
	defer s.Unlock()
	s.conn = c
}

func (s *wsSession) start() {
	atomic.StoreInt64(&s.closeInt, 0)
	// 重置发送队列
	s.sendQueueMaxLen = sendQueueMaxLen
	// todo 暂时默认发送队列长度2000
	s.sendQueue = make(chan interface{}, s.sendQueueMaxLen+1)

	s.exitWg.Add(2)
	s.node.(socket.SessionManager).Add(s)

	go func() {
		s.exitWg.Wait()
		// 结束操作处理
		close(s.sendQueue)

		s.node.(socket.SessionManager).Remove(s)
		if s.endCallback != nil {
			s.endCallback()
		}
	}()

	go s.RunRcv()
	go s.RunSend()
}

func (s *wsSession) RunRcv() {

}

func (s *wsSession) RunSend() {

}

func newWebSocketSession(conn *websocket.Conn, node iface.INetNode, endCallback func()) *wsSession {
	session := &wsSession{
		conn:        conn,
		node:        node,
		endCallback: endCallback,
		NetProcessorRPC: node.(interface {
			GetRPC() *socket.NetProcessorRPC
		}).GetRPC(), //使用外层node的RPC处理接口
	}
	node.(socket.Option).CopyOpt(&session.sessionOpt)
	return session
}
