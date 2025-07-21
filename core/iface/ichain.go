package iface

import "github.com/ljhe/scream/router"

type IChain interface {
	Execute(*router.Wrapper) error
}
