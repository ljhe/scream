package socket

import (
	"common"
	"github.com/gorilla/websocket"
	"math"
	"net"
	"time"
)

type Option interface {
	MaxMsgLen() int
	SocketReadTimeout(c net.Conn, callback func())
	SocketWriteTimeout(c net.Conn, callback func())
	WSReadTimeout(c *websocket.Conn, callback func())
	WSWriteTimeout(c *websocket.Conn, callback func())
	CopyOpt(opt *NetTCPSocketOption)
}

type NetTCPSocketOption struct {
	readBufferSize  int
	writeBufferSize int
	readTimeout     time.Duration
	writeTimeout    time.Duration
	noDelay         bool
	maxMsgLen       int
}

func (no *NetTCPSocketOption) Init() {
	no.maxMsgLen = common.MsgMaxLen
}

// SocketOptWebSocket 拷贝监听socket的配置信息
func (no *NetTCPSocketOption) SocketOptWebSocket(c *websocket.Conn) {
	if conn, ok := c.UnderlyingConn().(*net.TCPConn); ok {
		conn.SetNoDelay(no.noDelay)
		conn.SetReadBuffer(no.readBufferSize)
		conn.SetWriteBuffer(no.writeBufferSize)
	}
}

func (no *NetTCPSocketOption) MaxMsgLen() int {
	return no.maxMsgLen
}

func (no *NetTCPSocketOption) SocketReadTimeout(c net.Conn, callback func()) {
	if no.readTimeout > 0 {
		c.SetReadDeadline(time.Now().Add(no.readTimeout))
		callback()
		c.SetReadDeadline(time.Time{})
	} else {
		callback()
	}
}

func (no *NetTCPSocketOption) SocketWriteTimeout(c net.Conn, callback func()) {
	if no.writeTimeout > 0 {
		c.SetWriteDeadline(time.Now().Add(no.writeTimeout))
		callback()
		c.SetWriteDeadline(time.Time{})
	} else {
		callback()
	}
}

func (no *NetTCPSocketOption) WSReadTimeout(c *websocket.Conn, callback func()) {
	if no.readTimeout > 0 {
		c.SetReadDeadline(time.Now().Add(no.readTimeout))
		callback()
		c.SetReadDeadline(time.Time{})
	} else {
		callback()
	}
}

func (no *NetTCPSocketOption) WSWriteTimeout(c *websocket.Conn, callback func()) {
	if no.writeTimeout > 0 {
		c.SetWriteDeadline(time.Now().Add(no.writeTimeout))
		callback()
		c.SetWriteDeadline(time.Time{})
	} else {
		callback()
	}
}

func (no *NetTCPSocketOption) CopyOpt(opt *NetTCPSocketOption) {
	opt.maxMsgLen = no.maxMsgLen
	opt.noDelay = no.noDelay
	opt.readBufferSize = no.readBufferSize
	opt.writeBufferSize = no.writeBufferSize
}

func (no *NetTCPSocketOption) SetSocketBuff(read, write int, noDelay bool) {
	no.readBufferSize = read
	no.writeBufferSize = write
	no.noDelay = noDelay
	if read > 0 {
		no.maxMsgLen = read
	}
	if no.maxMsgLen >= math.MaxUint16 {
		no.maxMsgLen = math.MaxUint16
	}
}

func (no *NetTCPSocketOption) SetSocketDeadline(read, write time.Duration) {
	no.readTimeout = read
	no.writeTimeout = write
}
