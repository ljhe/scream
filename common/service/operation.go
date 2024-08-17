package service

import "fmt"

// SessionConnected 连接成功事件
type SessionConnected struct {
}

func (sc *SessionConnected) String() string {
	return fmt.Sprintf("%+v", *sc)
}

// SessionAccepted 接收其他服务器的连接
type SessionAccepted struct {
}

func (sa *SessionAccepted) String() string {
	return fmt.Sprintf("%+v", *sa)
}
