package grpc

import (
	"context"
	"github.com/ljhe/scream/core/iface"
	"github.com/ljhe/scream/pbgo"
	"google.golang.org/grpc"
	"log"
	"time"
)

type grpcConnector struct{}

func NewGRPCConnector() *grpcConnector {
	return &grpcConnector{}
}

func (g grpcConnector) Start() iface.INetNode {
	conn, err := grpc.Dial("localhost:9090", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := pbgo.NewAcceptorClient(conn)
	resp, err := client.Routing(context.Background(), &pbgo.RouteReqs{Msg: &pbgo.Message{
		Header: &pbgo.Header{
			Timestamp: time.Now().Unix(),
		},
		Body: []byte("grpc ping"),
	}})
	if err != nil {
		log.Fatalf("client.Routing: %v", err)
	}
	log.Printf("Greeting: %s", resp.Msg)
	return g
}

func (g grpcConnector) Stop() {
	//TODO implement me
	panic("implement me")
}

func (g grpcConnector) GetTyp() string {
	//TODO implement me
	panic("implement me")
}
