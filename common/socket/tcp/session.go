package tcp

import (
	"common/socket"
	"net"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"
)

var sendQueueMaxLen = 2000

type tcpSession struct {
	*socket.NetProcessorRPC
	conn            net.Conn
	close           int64
	sendQueue       chan interface{}
	sendQueueMaxLen int
	exitWg          sync.WaitGroup
	mu              sync.Mutex
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

func (ts *tcpSession) Start() {
	atomic.StoreInt64(&ts.close, 0)
	ts.sendQueueMaxLen = sendQueueMaxLen
	ts.sendQueue = make(chan interface{}, ts.sendQueueMaxLen)
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
