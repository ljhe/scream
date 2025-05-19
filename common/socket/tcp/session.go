package tcp

import (
	"github.com/ljhe/scream/common/iface"
	"github.com/ljhe/scream/common/socket"
	"github.com/ljhe/scream/plugins/logrus"
	"log"
	"net"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"
)

var sendQueueMaxLen = 200

type session struct {
	*socket.Processor // 事件处理相关
	socket.ContextSet // 记录session绑定信息
	node              iface.INetNode
	conn              net.Conn
	close             int64
	sendQueue         chan interface{}
	sendQueueMaxLen   int
	sessionOpt        socket.TCPSocketOption
	exitWg            sync.WaitGroup
	id                uint64
	rcvPingNum        int
	children          sync.Map
	mu                sync.Mutex
}

func (ts *session) SetConn(c net.Conn) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.conn = c
}

func (ts *session) GetConn() net.Conn {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.conn
}

func (ts *session) SetId(id uint64) {
	ts.id = id
}

func (ts *session) GetId() uint64 {
	return ts.id
}

func (ts *session) Raw() interface{} {
	return ts.GetConn()
}

func (ts *session) Node() iface.INetNode {
	return ts.node
}

func (ts *session) IncRcvPingNum(inc int) {
	if inc <= 0 {
		ts.rcvPingNum = inc
		return
	}
	ts.rcvPingNum += inc
}

func (ts *session) RcvPingNum() int {
	return ts.rcvPingNum
}

func NewTcpSession(c net.Conn, node iface.INetNode) *session {
	sess := &session{
		conn:            c,
		node:            node,
		sendQueueMaxLen: sendQueueMaxLen,
		sendQueue:       make(chan interface{}, sendQueueMaxLen),
		Processor: node.(interface {
			GetRPC() *socket.Processor
		}).GetRPC(),
	}
	node.(socket.Option).CopyOpt(&sess.sessionOpt)
	return sess
}

func (ts *session) Start() {
	atomic.StoreInt64(&ts.close, 0)
	ts.exitWg.Add(2)
	// 添加到session管理器中
	ts.node.(iface.ISessionManager).Add(ts)
	go func() {
		ts.exitWg.Wait()
		close(ts.sendQueue)
		ts.node.(iface.ISessionManager).Remove(ts)
	}()
	go ts.RunRcv()
	go ts.RunSend()
}

func (ts *session) Close() {
	// 已经关闭
	if ok := atomic.SwapInt64(&ts.close, 1); ok != 0 {
		return
	}
	conn := ts.GetConn()
	if conn != nil {
		conn.Close()
	}
}

func (ts *session) RunRcv() {
	defer func() {
		if err := recover(); err != nil {
			logrus.Log(logrus.LogsSystem).Errorf("tcpSession Stack---::%v\n %s\n", err, string(debug.Stack()))
			debug.PrintStack()
		}
	}()

	for {
		msg, err := ts.ReadMsg(ts)
		if err != nil {
			logrus.Log(logrus.LogsSystem).Errorf("RunRcv ReadMsg err:%v sessionId:%d \n", err, ts.GetId())
			// 做关闭处理 发送数据时已经无法发送
			atomic.StoreInt64(&ts.close, 1)
			select {
			case ts.sendQueue <- nil:
			default:
				logrus.Log(logrus.LogsSystem).Errorf("RunRcv sendQueue block len:%d sessionId:%d \n", len(ts.sendQueue), ts.GetId())
			}

			// 抛出关闭事件
			ts.ProcEvent(&socket.RcvMsgEvent{Sess: ts, Message: &socket.SessionClosed{}, Err: err})
			break
		}
		ts.ProcEvent(&socket.RcvMsgEvent{Sess: ts, Message: msg})
	}
	ts.exitWg.Done()
}

func (ts *session) RunSend() {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("Stack---::%v\n %s\n", err, string(debug.Stack()))
		}
	}()

	for data := range ts.sendQueue {
		if atomic.LoadInt64(&ts.close) == 1 {
			break
		}
		if data == nil {
			continue
		}
		err := ts.SendMsg(&socket.SendMsgEvent{Sess: ts, Message: data})
		if err != nil {
			log.Println("send msg err:", err)
			break
		}
	}

	conn := ts.GetConn()
	if conn != nil {
		conn.Close()
	}
	ts.exitWg.Done()
}

// HeartBeat 服务器之间的心跳检测
func (ts *session) HeartBeat(msg interface{}) {
	if atomic.LoadInt64(&ts.close) != 0 {
		return
	}

	go func() {
		delayTimer := time.NewTimer(15 * time.Second)
		for {
			delayTimer.Reset(5 * time.Second)
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

func (ts *session) Send(msg interface{}) {
	if atomic.LoadInt64(&ts.close) != 0 {
		return
	}
	select {
	case ts.sendQueue <- msg:
	default:
	}
}

func (ts *session) SetSessionChild(sessionId uint64, data interface{}) {
	var sc *SessionChild
	if val, ok := ts.children.Load(sessionId); !ok {
		sc = NewSessionChild(sessionId, ts)
		sc.Start()
		ts.children.Store(sessionId, sc)
	} else {
		sc = val.(*SessionChild)
	}
	sc.Rcv(data)
}

func (ts *session) DelSessionChild(sessionId uint64) {
	if val, ok := ts.children.Load(sessionId); ok {
		val.(*SessionChild).Stop()
		ts.children.Delete(sessionId)
	}
}
