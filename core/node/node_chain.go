package node

import "github.com/ljhe/scream/router"

type EventHandler func(*router.Wrapper) error

type DefaultChain struct {
	Before  []EventHandler
	After   []EventHandler
	Handler EventHandler
}

func (c *DefaultChain) Execute(mw *router.Wrapper) error {
	var err error

	for _, before := range c.Before {
		err = before(mw)
		if err != nil {
			goto ext
		}
	}

	err = c.Handler(mw)
	if err != nil {
		goto ext
	}

	for _, after := range c.After {
		err = after(mw)
		if err != nil {
			goto ext
		}
	}

ext:
	return err
}
