package tests

import (
	"context"
	"fmt"
	"github.com/alicebob/miniredis/v2"
	"github.com/ljhe/scream/3rd/etcd"
	log "github.com/ljhe/scream/3rd/log"
	"github.com/ljhe/scream/3rd/redis"
	"github.com/ljhe/scream/core/iface"
	"github.com/ljhe/scream/core/node"
	"github.com/ljhe/scream/core/process"
	"github.com/ljhe/scream/tests/mock"
	"os"
	"testing"
	"time"
)

var factory *mock.NodeFactory
var loader iface.INodeLoader

func TestMain(m *testing.M) {
	logger, err := log.NewDefaultLogger()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	etcd.InitServiceDiscovery("127.0.0.1:2379")

	factory = mock.BuildNodeFactory()
	loader = node.BuildDefaultNodeLoader(factory)

	mr, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer mr.Close()
	redis.BuildClientWithOption(redis.WithAddr(fmt.Sprintf("redis://%s", mr.Addr())))

	registerMockA()

	os.Exit(m.Run())
}

var p1 iface.IProcess

func registerMockA() {
	p1 = process.BuildProcessWithOption(
		process.WithID("bench-call-1"),
		process.WithPort(8888),
		process.WithLoader(loader),
		process.WithFactory(factory),
	)

	// build
	p1.System().Loader("mocka").WithID("mocka").Register(context.TODO())

	p1.Init()

	fmt.Println("register test process ok")
	time.Sleep(time.Second)
}
