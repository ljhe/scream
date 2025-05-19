package websocket

import (
	"github.com/gorilla/websocket"
	"github.com/ljhe/scream/common/iface"
	"github.com/ljhe/scream/common/socket"
	"github.com/ljhe/scream/pbgo"
	"github.com/ljhe/scream/plugins/logrus"
	"runtime/debug"
	"sync"
	"sync/atomic"
)

var sendQueueMaxLen = 200

type session struct {
	sync.Mutex
	*socket.Processor // 事件处理相关 processor.go
	socket.ContextSet // 记录session绑定信息 nodeproperty.go
	node              iface.INetNode
	conn              *websocket.Conn

	exitWg      sync.WaitGroup
	id          uint64
	endCallback func()
	close       int64

	sessionOpt socket.TCPSocketOption

	sendQueue       chan interface{} // 消息发送队列
	sendQueueMaxLen int
}

func (s *session) SetConn(c *websocket.Conn) {
	s.Lock()
	defer s.Unlock()
	s.conn = c
}

func (s *session) GetConn() *websocket.Conn {
	s.Lock()
	defer s.Unlock()
	return s.conn
}

func (s *session) Raw() interface{} {
	return s.GetConn()
}

func (s *session) Node() iface.INetNode {
	return s.node
}

func (s *session) Send(msg interface{}) {
	if atomic.LoadInt64(&s.close) != 0 {
		return
	}
	select {
	case s.sendQueue <- msg:
	default:
		logrus.Log(logrus.LogsSystem).Errorf("SendLen-sendQueue block len=%d sessionId=%d addr=%v", len(s.sendQueue), s.GetId(), s.conn.LocalAddr())
	}
}

func (s *session) Close() {
	//已经关闭
	if ok := atomic.SwapInt64(&s.close, 1); ok != 0 {
		return
	}

	conn := s.GetConn()
	if conn != nil {
		//conn.Close()
		//关闭读
		conn.Close()
		conn.CloseHandler()
	}
}

func (s *session) SetId(id uint64) {
	s.id = id
}

func (s *session) GetId() uint64 {
	return s.id
}

func (s *session) HeartBeat(msg interface{}) {

}

func (s *session) IncRcvPingNum(inc int) {

}

func (s *session) RcvPingNum() int {
	return 0
}

func (s *session) setConn(c *websocket.Conn) {
	s.Lock()
	defer s.Unlock()
	s.conn = c
}

func (s *session) Start() {
	atomic.StoreInt64(&s.close, 0)

	s.sendQueueMaxLen = sendQueueMaxLen
	s.sendQueue = make(chan interface{}, s.sendQueueMaxLen+1)

	s.exitWg.Add(2)
	s.node.(iface.ISessionManager).Add(s)

	go func() {
		s.exitWg.Wait()
		// 结束操作处理
		close(s.sendQueue)

		s.node.(iface.ISessionManager).Remove(s)
		if s.endCallback != nil {
			s.endCallback()
		}
	}()

	go s.RunRcv()
	go s.RunSend()
}

func (s *session) RunRcv() {
	defer func() {
		// 打印堆栈信息
		if err := recover(); err != nil {
			logrus.Log(logrus.LogsSystem).Errorf("wsSession Stack---::%v\n %s\n", err, string(debug.Stack()))
			debug.PrintStack()
		}
	}()

	for {
		msg, err := s.ReadMsg(s)
		if err != nil {
			// 检测是否正常关闭
			if !websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				logrus.Log(logrus.LogsSystem).Errorf("RunRcv ReadMsg err:%v sessionId:%d \n", err, s.GetId())
			}
			// 做关闭处理 发送数据时已经无法发送
			atomic.StoreInt64(&s.close, 1)
			select {
			case s.sendQueue <- nil:
			default:
				logrus.Log(logrus.LogsSystem).Errorf("RunRcv sendQueue block len:%d sessionId:%d \n", len(s.sendQueue), s.GetId())
			}

			// 抛出关闭事件
			s.ProcEvent(&socket.RcvMsgEvent{Sess: s, Message: &pbgo.WSSessionClosedNtf{}, Err: err})
			break
		}

		// 接收数据事件放到队列中(需要放到队列中，否则会有线程冲突)
		s.ProcEvent(&socket.RcvMsgEvent{Sess: s, Message: msg, Err: nil})
	}

	logrus.Log(logrus.LogsSystem).Infof("wsSession exit addr:%v", s.conn.LocalAddr())
	s.exitWg.Done()
}

func (s *session) RunSend() {
	defer func() {
		// 打印堆栈信息
		if err := recover(); err != nil {
			logrus.Log(logrus.LogsSystem).Errorf("wsSession Stack---::%v\n %s\n", err, string(debug.Stack()))
			debug.PrintStack()
		}
	}()

	for data := range s.sendQueue {
		if data == nil {
			break
		}
		err := s.SendMsg(&socket.SendMsgEvent{Sess: s, Message: data})
		if err != nil {
			logrus.Log(logrus.LogsSystem).Errorf("wsSession RunSend SendMsg err:%v \n", err)
			break
		}
	}

	logrus.Log(logrus.LogsSystem).Infof("wsSession RunSend exit RunSend goroutine addr=%v", s.conn.LocalAddr())
	c := s.GetConn()
	if c != nil {
		c.Close()
	}

	s.exitWg.Done()
}

func (s *session) SetSessionChild(sessionId uint64, data interface{}) {

}

func (s *session) DelSessionChild(sessionId uint64) {
	
}

func newWSSession(conn *websocket.Conn, node iface.INetNode, endCallback func()) *session {
	sess := &session{
		conn:        conn,
		node:        node,
		endCallback: endCallback,
		// 在session中初始化 每一个session 一个处理消息的队列 实现多客户端之间并行
		Processor: &socket.Processor{
			MsgProc: new(socket.WSMessageProcessor),
			Hooker:  new(socket.WsHookEvent),
		},
	}
	node.(socket.Option).CopyOpt(&sess.sessionOpt)
	return sess
}
