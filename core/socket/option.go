package socket

import (
	"github.com/gorilla/websocket"
	"github.com/ljhe/scream/core/iface"
	"github.com/ljhe/scream/def"
	"net"
	"time"
)

type Option struct {
	ReadBufferSize  int
	WriteBufferSize int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	MaxMsgLen       int
}

func (o *Option) GetMaxMsgLen() int {
	return o.MaxMsgLen
}

func (o *Option) SocketReadTimeout(s iface.ISession, callback func()) {
	switch s.Conn().(type) {
	case net.Conn:
		if o.ReadTimeout > 0 {
			callback()
			s.Conn().(net.Conn).SetReadDeadline(time.Now().Add(o.ReadTimeout))
		} else {
			callback()
		}
	case *websocket.Conn:
		if o.ReadTimeout > 0 {
			callback()
			s.Conn().(*websocket.Conn).SetReadDeadline(time.Now().Add(o.ReadTimeout))
		} else {
			callback()
		}
	}
}

func (o *Option) SocketWriteTimeout(s iface.ISession, callback func()) {
	switch s.Conn().(type) {
	case net.Conn:
		if o.ReadTimeout > 0 {
			callback()
			s.Conn().(net.Conn).SetWriteDeadline(time.Now().Add(o.WriteTimeout))
		} else {
			callback()
		}
	case *websocket.Conn:
		if o.ReadTimeout > 0 {
			callback()
			s.Conn().(*websocket.Conn).SetWriteDeadline(time.Now().Add(o.WriteTimeout))
		} else {
			callback()
		}
	}
}

func (o *Option) SetOption(option interface{}) {
	// 默认是最大值
	o.MaxMsgLen = def.MsgMaxLen

	if opt, ok := option.(*Option); ok {
		o.ReadBufferSize = opt.ReadBufferSize
		o.WriteBufferSize = opt.WriteBufferSize
		o.ReadTimeout = opt.ReadTimeout
		o.WriteTimeout = opt.WriteTimeout
		if opt.MaxMsgLen > 0 {
			o.MaxMsgLen = opt.MaxMsgLen
		}
	}
}
