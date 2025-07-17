package grpc

import (
	"context"
	"github.com/ljhe/scream/3rd/logrus"
	"github.com/ljhe/scream/core/iface"
	"github.com/ljhe/scream/message"
	"google.golang.org/grpc"
	"log"
	"net"
)

type grpcAcceptor struct {
	*Runtime
}

type listen struct {
	message.AcceptorServer
	g *grpcAcceptor
}

func NewGRPCAcceptor() *grpcAcceptor {
	return &grpcAcceptor{
		Runtime: &Runtime{},
	}
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
	message.RegisterAcceptorServer(gs, &listen{g: g})
	if err := gs.Serve(ln); err != nil {
		panic(err)
	}
	log.Println("gRPC server listening success.")
	return g
}

func (l *listen) Routing(ctx context.Context, req *message.RouteReq) (*message.RouteRes, error) {
	res := &message.RouteRes{
		Msg: &message.Message{},
	}

	err := l.g.Received(req)
	if err != nil {
		logrus.Errorf("groc Routing received err:%v", err)
		return nil, err
	}

	res.Msg.Body = []byte("return pong")
	return res, nil
}
