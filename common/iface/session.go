package iface

type ISession interface {
	Send(msg interface{})
	HeartBeat(msg interface{})
}
