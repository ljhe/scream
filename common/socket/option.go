package socket

import (
	"common"
	"github.com/gorilla/websocket"
	"net"
	"time"
)

type Option interface {
	SocketReadTimeout(c net.Conn, callback func())
	SocketWriteTimeout(c net.Conn, callback func())
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

func (no *NetTCPSocketOption) CopyOpt(opt *NetTCPSocketOption) {
	opt.maxMsgLen = no.maxMsgLen
	opt.noDelay = no.noDelay
	opt.readBufferSize = no.readBufferSize
	opt.writeBufferSize = no.writeBufferSize
}
