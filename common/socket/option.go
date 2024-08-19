package socket

import (
	"net"
	"time"
)

type Option interface {
	SocketWriteTimeout(c net.Conn, callback func())
}

type NetTCPSocketOption struct {
	writeTimeout time.Duration
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
