package grpc

import (
	"context"
	"github.com/ljhe/scream/core/iface"
	"github.com/ljhe/scream/message"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"time"
)

type grpcConnector struct{}

func NewGRPCConnector() *grpcConnector {
	return &grpcConnector{}
}

func (g grpcConnector) Start() iface.INetNode {
	conn, err := grpc.NewClient(
		"localhost:9090",
		// 测试环境使用的是明文连接 生产环境最好用成TLS安全连接
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := message.NewAcceptorClient(conn)
	resp, err := client.Routing(context.Background(), &message.RouteReq{Msg: &message.Message{
		Header: &message.Header{
			Event:     "ping",
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
