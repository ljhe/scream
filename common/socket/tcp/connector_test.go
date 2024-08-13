package tcp

import (
	"common"
	"common/service"
	"common/socket"
	"math/rand"
	"testing"
	"time"
)

func TestNewConnector(t *testing.T) {
	node := socket.NewServerNode(common.SocketTypTcpConnector, "test", "0.0.0.0:2701")
	rand.New(rand.NewSource(time.Now().UnixNano()))
	node.Start()
	service.WaitExitSignal()
}
