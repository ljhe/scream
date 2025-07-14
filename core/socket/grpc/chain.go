package grpc

type DefaultChain struct {
	Handler func() error
}

func (c *DefaultChain) Execute() error {
	c.Handler()
	return nil
}
