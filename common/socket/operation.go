package socket

import (
	_ "pbgo"
)

// SessionConnected 连接成功事件
type SessionConnected struct {
}

// SessionAccepted 接收其他服务器的连接
type SessionAccepted struct {
}

// SessionClosed 连接关闭事件
type SessionClosed struct {
}
