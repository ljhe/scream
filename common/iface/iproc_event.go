package iface

type EventCallBack func(e IProcEvent)

type IProcEvent interface {
	Session() ISession
	Msg() interface{}
}
