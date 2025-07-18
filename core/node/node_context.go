package node

import (
	"context"
	"errors"
	"github.com/ljhe/scream/core/iface"
)

type systemKey struct{}

type Context struct {
	ctx context.Context
}

func (nc *Context) Unregister(id, ty string) error {
	sys, ok := nc.ctx.Value(systemKey{}).(iface.ISystem)
	if !ok {
		panic(errors.New("the system instance does not exist in the NodeContext"))
	}

	return sys.Unregister(id, ty)
}
