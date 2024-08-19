package iface

type IProcEvent interface {
	Session() ISession
	Msg() interface{}
}
