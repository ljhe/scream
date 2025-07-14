package tests

import (
	"context"
	"fmt"
	"github.com/ljhe/scream/core/iface"
	"github.com/ljhe/scream/core/socket/grpc"
	"testing"
)

func TestGRPCAcceptor(t *testing.T) {
	g := grpc.NewGRPCAcceptor()

	g.Init(context.TODO())
	g.OnEvent("ping", func() iface.IChain {
		return &grpc.DefaultChain{
			Handler: func() error {
				fmt.Println("received ping")
				return nil
			},
		}
	})

	g.Start()
}

func TestGRPCConnector(t *testing.T) {
	g := grpc.NewGRPCConnector()
	g.Start()
}
