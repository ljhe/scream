package iface

import "github.com/ljhe/scream/msg"

type IFuture interface {
	Complete(*msg.Wrapper)
	IsCompleted() bool

	Then(func(*msg.Wrapper)) IFuture
}
