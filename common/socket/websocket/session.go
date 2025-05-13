package websocket

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/ljhe/scream/common/iface"
	"github.com/ljhe/scream/common/socket"
	"github.com/ljhe/scream/plugins/logrus"
	"runtime/debug"
	"sync"
	"sync/atomic"
)

var sendQueueMaxLen = 2000
var rcvQueueMaxLen = 2000

type wsSession struct {
	sync.Mutex
	*socket.NetProcessorRPC // 事件处理相关 procrpc.go
	socket.NetContextSet    // 记录session绑定信息 nodeproperty.go
	node                    iface.INetNode
	conn                    *websocket.Conn

	exitWg      sync.WaitGroup
	id          uint64
	endCallback func()
	close       int64

	sessionOpt socket.NetTCPSocketOption

	sendQueue       chan interface{} // 消息发送队列
	rcvQueue        chan interface{} // 消息接收队列
	sendQueueMaxLen int
	rcvQueueMaxLen  int
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
	if atomic.LoadInt64(&s.close) != 0 {
		return
	}
	select {
	case s.sendQueue <- msg:
	default:
		logrus.Log(logrus.LogsSystem).Errorf("SendLen-sendQueue block len=%d sessionId=%d addr=%v", len(s.sendQueue), s.GetId(), s.conn.LocalAddr())
	}
}

func (s *wsSession) Close() {
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

func (s *wsSession) SetId(id uint64) {
	s.id = id
}

func (s *wsSession) GetId() uint64 {
	return s.id
}

func (s *wsSession) HeartBeat(msg interface{}) {

}

func (s *wsSession) IncRcvPingNum(inc int) {

}

func (s *wsSession) RcvPingNum() int {
	return 0
}

func (s *wsSession) setConn(c *websocket.Conn) {
	s.Lock()
	defer s.Unlock()
	s.conn = c
}

func (s *wsSession) start() {
	atomic.StoreInt64(&s.close, 0)

	s.sendQueueMaxLen = sendQueueMaxLen
	s.rcvQueueMaxLen = rcvQueueMaxLen
	// todo 暂时默认发送 接收队列长度2000
	s.sendQueue = make(chan interface{}, s.sendQueueMaxLen+1)
	s.rcvQueue = make(chan interface{}, s.rcvQueueMaxLen+1)

	s.exitWg.Add(3)
	s.node.(socket.SessionManager).Add(s)

	go func() {
		s.exitWg.Wait()
		// 结束操作处理
		close(s.sendQueue)
		close(s.rcvQueue)

		s.node.(socket.SessionManager).Remove(s)
		if s.endCallback != nil {
			s.endCallback()
		}
	}()

	go s.RunRcv()
	go s.RunSend()
	go s.RunRcvQueue()
}

func (s *wsSession) RunRcv() {
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
			logrus.Log(logrus.LogsSystem).Errorf("RunRcv ReadMsg err:%v sessionId:%d \n", err, s.GetId())
			// 做关闭处理 发送数据时已经无法发送
			atomic.StoreInt64(&s.close, 1)
			select {
			case s.sendQueue <- nil:
			default:
				logrus.Log(logrus.LogsSystem).Errorf("RunRcv sendQueue block len:%d sessionId:%d \n", len(s.sendQueue), s.GetId())
			}

			// 抛出关闭事件
			//s.ProcEvent(&socket.RcvMsgEvent{Sess: s, Message: &socket.SessionClosed{}, Err: err})
			s.rcvQueue <- nil
			break
		}

		// 接收数据事件放到队列中(需要放到队列中，否则会有线程冲突)
		// 单线程
		//s.ProcEvent(&socket.RcvMsgEvent{Sess: s, Message: msg, Err: nil})
		// 多线程
		s.rcvQueue <- msg
	}

	logrus.Log(logrus.LogsSystem).Infof("wsSession exit addr:%v", s.conn.LocalAddr())
	s.exitWg.Done()
}

func (s *wsSession) RunSend() {
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

func (s *wsSession) RunRcvQueue() {
	defer func() {
		// 打印堆栈信息
		if err := recover(); err != nil {
			logrus.Log(logrus.LogsSystem).Errorf("wsSession Stack---::%v\n %s\n", err, string(debug.Stack()))
			debug.PrintStack()
		}
	}()

	for data := range s.rcvQueue {
		if data == nil {
			break
		}
		fmt.Println(fmt.Sprintf("rcv queue data:%v \n", data))
		// TODO 根据消息的不同类型来转发到不同的地方
		//err := s.SendMsg(&socket.SendMsgEvent{Sess: s, Message: data})
		//if err != nil {
		//	logrus.Log(logrus.LogsSystem).Errorf("wsSession RunSend SendMsg err:%v \n", err)
		//	break
		//}
	}

	logrus.Log(logrus.LogsSystem).Infof("wsSession RunRcvQueue exit. sessionId=%d", s.id)
	s.exitWg.Done()
}

func newWebSocketSession(conn *websocket.Conn, node iface.INetNode, endCallback func()) *wsSession {
	session := &wsSession{
		conn:        conn,
		node:        node,
		endCallback: endCallback,
		NetProcessorRPC: node.(interface {
			GetRPC() *socket.NetProcessorRPC
		}).GetRPC(), // 使用外层node的RPC处理接口
	}
	node.(socket.Option).CopyOpt(&session.sessionOpt)
	return session
}
