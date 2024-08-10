package tcp

import (
	"common"
	"common/service"
	"common/socket"
	"testing"
)

func TestNewAcceptor(t *testing.T) {
	node := socket.NewServerNode(common.SocketTypTcpAcceptor, "test", "0.0.0.0:2701")
	node.Start()
	service.WaitExitSignal()
}
