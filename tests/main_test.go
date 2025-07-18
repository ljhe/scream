package tests

import (
	"github.com/ljhe/scream/3rd/etcd"
	"github.com/ljhe/scream/3rd/logrus"
	"github.com/ljhe/scream/core/iface"
	"github.com/ljhe/scream/core/node"
	"github.com/ljhe/scream/tests/mock"
	"os"
	"testing"
)

var factory *mock.NodeFactory
var loader iface.INodeLoader

func TestMain(m *testing.M) {
	logrus.Init("")
	etcd.InitServiceDiscovery("127.0.0.1:2379")

	factory = mock.BuildNodeFactory()
	loader = node.BuildDefaultActorLoader(factory)

	os.Exit(m.Run())
}
