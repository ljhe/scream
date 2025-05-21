package socket

import (
	"github.com/gorilla/websocket"
	"github.com/ljhe/scream/core"
	"github.com/ljhe/scream/core/iface"
	"math"
	"net"
	"time"
)

type Option interface {
	MaxMsgLen() int
	SocketReadTimeout(s iface.ISession, callback func())
	SocketWriteTimeout(s iface.ISession, callback func())
	CopyOpt(opt *TCPSocketOption)
}

type TCPSocketOption struct {
	readBufferSize  int
	writeBufferSize int
	readTimeout     time.Duration
	writeTimeout    time.Duration
	noDelay         bool
	maxMsgLen       int
}

func (no *TCPSocketOption) Init() {
	no.maxMsgLen = core.MsgMaxLen
}

// SocketOptWebSocket 拷贝监听socket的配置信息
func (no *TCPSocketOption) SocketOptWebSocket(c *websocket.Conn) {
	if conn, ok := c.UnderlyingConn().(*net.TCPConn); ok {
		conn.SetNoDelay(no.noDelay)
		conn.SetReadBuffer(no.readBufferSize)
		conn.SetWriteBuffer(no.writeBufferSize)
	}
}

func (no *TCPSocketOption) MaxMsgLen() int {
	return no.maxMsgLen
}

func (no *TCPSocketOption) SocketReadTimeout(s iface.ISession, callback func()) {
	switch s.Conn().(type) {
	case net.Conn:
		if no.readTimeout > 0 {
			s.Conn().(net.Conn).SetReadDeadline(time.Now().Add(no.readTimeout))
			callback()
			s.Conn().(net.Conn).SetReadDeadline(time.Time{})
		} else {
			callback()
		}
	case *websocket.Conn:
		if no.readTimeout > 0 {
			s.Conn().(*websocket.Conn).SetReadDeadline(time.Now().Add(no.readTimeout))
			callback()
			s.Conn().(*websocket.Conn).SetReadDeadline(time.Time{})
		} else {
			callback()
		}
	}
}

func (no *TCPSocketOption) SocketWriteTimeout(s iface.ISession, callback func()) {
	switch s.Conn().(type) {
	case net.Conn:
		if no.readTimeout > 0 {
			s.Conn().(net.Conn).SetWriteDeadline(time.Now().Add(no.readTimeout))
			callback()
			s.Conn().(net.Conn).SetWriteDeadline(time.Time{})
		} else {
			callback()
		}
	case *websocket.Conn:
		if no.readTimeout > 0 {
			s.Conn().(*websocket.Conn).SetWriteDeadline(time.Now().Add(no.readTimeout))
			callback()
			s.Conn().(*websocket.Conn).SetWriteDeadline(time.Time{})
		} else {
			callback()
		}
	}
}

func (no *TCPSocketOption) WSReadTimeout(c *websocket.Conn, callback func()) {
	if no.readTimeout > 0 {
		c.SetReadDeadline(time.Now().Add(no.readTimeout))
		callback()
		c.SetReadDeadline(time.Time{})
	} else {
		callback()
	}
}

func (no *TCPSocketOption) WSWriteTimeout(c *websocket.Conn, callback func()) {
	if no.writeTimeout > 0 {
		c.SetWriteDeadline(time.Now().Add(no.writeTimeout))
		callback()
		c.SetWriteDeadline(time.Time{})
	} else {
		callback()
	}
}

func (no *TCPSocketOption) CopyOpt(opt *TCPSocketOption) {
	opt.maxMsgLen = no.maxMsgLen
	opt.noDelay = no.noDelay
	opt.readBufferSize = no.readBufferSize
	opt.writeBufferSize = no.writeBufferSize
}

func (no *TCPSocketOption) SetSocketBuff(read, write int, noDelay bool) {
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

func (no *TCPSocketOption) SetSocketDeadline(read, write time.Duration) {
	no.readTimeout = read
	no.writeTimeout = write
}
