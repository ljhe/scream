package iface

import "github.com/ljhe/scream/msg"

type IChain interface {
	Execute(*msg.Wrapper) error
}
