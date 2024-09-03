package tcp

import (
	"common"
	"common/iface"
	"common/socket"
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
	node                    iface.INetNode
	conn                    net.Conn
	close                   int64
	sendQueue               chan interface{}
	sendQueueMaxLen         int
	sessionOpt              socket.NetTCPSocketOption
	exitWg                  sync.WaitGroup
	id                      uint64
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

func (ts *tcpSession) Node() iface.INetNode {
	return ts.node
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
			debug.PrintStack()
		}
	}()

	for {
		msg, err := ts.ReadMsg(ts)
		if err != nil {
			log.Printf("RunRcv ReadMsg err:%v sessionId:%d \n", err, ts.GetId())
			break
		}
		ts.ProcEvent(&common.RcvMsgEvent{Sess: ts, Message: msg})
	}
	ts.exitWg.Done()
}

func (ts *tcpSession) RunSend() {
	defer func() {
		if err := recover(); err != nil {
			debug.PrintStack()
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
		log.Printf("send msg: %v \n", data)
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
