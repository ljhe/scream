package tcp

import (
	"github.com/ljhe/scream/common"
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

var sendQueueMaxLen = 2000

type tcpSession struct {
	*socket.NetProcessorRPC // 事件处理相关
	socket.NetContextSet    // 记录session绑定信息
	node                    iface.INetNode
	conn                    net.Conn
	close                   int64
	sendQueue               chan interface{}
	sendQueueMaxLen         int
	sessionOpt              socket.NetTCPSocketOption
	exitWg                  sync.WaitGroup
	id                      uint64
	rcvPingNum              int
	mu                      sync.Mutex
}

func (ts *tcpSession) SetConn(c net.Conn) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.conn = c
}

func (ts *tcpSession) GetConn() net.Conn {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.conn
}

func (ts *tcpSession) SetId(id uint64) {
	ts.id = id
}

func (ts *tcpSession) GetId() uint64 {
	return ts.id
}

func (ts *tcpSession) Raw() interface{} {
	return ts.GetConn()
}

func (ts *tcpSession) Node() iface.INetNode {
	return ts.node
}

func (ts *tcpSession) IncRcvPingNum(inc int) {
	if inc <= 0 {
		ts.rcvPingNum = inc
		return
	}
	ts.rcvPingNum += inc
}

func (ts *tcpSession) RcvPingNum() int {
	return ts.rcvPingNum
}

func newTcpSession(c net.Conn, node iface.INetNode) *tcpSession {
	sess := &tcpSession{
		conn:            c,
		node:            node,
		sendQueueMaxLen: sendQueueMaxLen,
		sendQueue:       make(chan interface{}, sendQueueMaxLen),
		NetProcessorRPC: node.(interface {
			GetRPC() *socket.NetProcessorRPC
		}).GetRPC(),
	}
	node.(socket.Option).CopyOpt(&sess.sessionOpt)
	return sess
}

func (ts *tcpSession) Start() {
	atomic.StoreInt64(&ts.close, 0)
	ts.exitWg.Add(2)
	// 添加到session管理器中
	ts.node.(socket.SessionManager).Add(ts)
	go func() {
		ts.exitWg.Wait()
		close(ts.sendQueue)
		ts.node.(socket.SessionManager).Remove(ts)
	}()
	go ts.RunRcv()
	go ts.RunSend()
}

func (ts *tcpSession) Close() {
	// 已经关闭
	if ok := atomic.SwapInt64(&ts.close, 1); ok != 0 {
		return
	}
	conn := ts.GetConn()
	if conn != nil {
		conn.Close()
	}
}

func (ts *tcpSession) RunRcv() {
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
			ts.ProcEvent(&common.RcvMsgEvent{Sess: ts, Message: &socket.SessionClosed{}, Err: err})
			break
		}
		ts.ProcEvent(&common.RcvMsgEvent{Sess: ts, Message: msg})
	}
	ts.exitWg.Done()
}

func (ts *tcpSession) RunSend() {
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
		err := ts.SendMsg(&common.SendMsgEvent{Sess: ts, Message: data})
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
func (ts *tcpSession) HeartBeat(msg interface{}) {
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

func (ts *tcpSession) Send(msg interface{}) {
	if atomic.LoadInt64(&ts.close) != 0 {
		return
	}
	select {
	case ts.sendQueue <- msg:
	default:
	}
}
