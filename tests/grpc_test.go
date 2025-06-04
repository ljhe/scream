package tests

import (
	"github.com/ljhe/scream/core/socket/grpc"
	"testing"
)

func TestGRPCAcceptor(t *testing.T) {
	gs := grpc.NewGRPCAcceptor()
	gs.Start()
}

func TestGRPCConnector(t *testing.T) {
	gc := grpc.NewGRPCConnector()
	gc.Start()
}
