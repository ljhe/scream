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

func (ts *tcpSession) Node() iface.INetNode {
	return ts.node
}

func newTcpSession(c net.Conn, node iface.INetNode) *tcpSession {
	sess := &tcpSession{
		conn:            c,
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
	go func() {
		ts.exitWg.Wait()
		close(ts.sendQueue)
	}()
	go ts.RunRcv()
	go ts.RunSend()
}

func (ts *tcpSession) RunRcv() {

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
		log.Println("send msg:", data)
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
