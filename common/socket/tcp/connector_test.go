package tcp

import (
	"common"
	"common/service"
	"common/socket"
	"testing"
)

func TestNewConnector(t *testing.T) {
	node := socket.NewServerNode(common.SocketTypTcpConnector, "test", "0.0.0.0:2701")
	msgHandle := service.GetMsgHandle(0)
	node.(common.ProcessorRPCBundle).SetHooker(new(service.ServerEventHook))
	node.(common.ProcessorRPCBundle).SetMsgHandle(msgHandle)
	node.Start()
	service.WaitExitSignal()
}
