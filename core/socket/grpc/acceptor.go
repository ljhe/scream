package grpc

import (
	"context"
	"fmt"
	"github.com/ljhe/scream/core/iface"
	"github.com/ljhe/scream/pbgo"
	"google.golang.org/grpc"
	"log"
	"net"
)

type grpcAcceptor struct {
}

type listen struct {
	pbgo.AcceptorServer
}

func NewGRPCAcceptor() *grpcAcceptor {
	return &grpcAcceptor{}
}

func (g *grpcAcceptor) Stop() {
	//TODO implement me
	panic("implement me")
}

func (g *grpcAcceptor) GetTyp() string {
	//TODO implement me
	panic("implement me")
}

func (g *grpcAcceptor) Start() iface.INetNode {
	ln, err := net.Listen("tcp", ":9090")
	if err != nil {
		panic(err)
	}

	gs := grpc.NewServer()
	pbgo.RegisterAcceptorServer(gs, &listen{})
	if err := gs.Serve(ln); err != nil {
		panic(err)
	}
	log.Println("gRPC server listening success.")
	return g
}

func (l *listen) Routing(ctx context.Context, req *pbgo.RouteReqs) (*pbgo.RouteRes, error) {
	fmt.Println("grpc pong")
	return nil, nil
}
