package service

import (
	"common/iface"
	"log"
)

type ServerEventHook struct {
}

func (eh *ServerEventHook) InEvent(iv iface.IProcEvent) iface.IProcEvent {
	switch msg := iv.Msg().(type) {
	case *SessionAccepted:
		log.Println("服务器连接成功111", msg)
		return nil
	case *SessionConnected:
		log.Println("服务器连接成功222", msg)
		return nil
	}
	return iv
}

func (eh *ServerEventHook) OutEvent(ov iface.IProcEvent) iface.IProcEvent {
	return ov
}
