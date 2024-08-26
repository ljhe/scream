package service

import (
	"common"
	"common/iface"
	plugins "common/plugins/etcd"
	"common/socket"
	"log"
	"os"
	"os/signal"
	"syscall"
)

type NetNodeParam struct {
	ServerTyp  string
	ServerName string
	Addr       string
	Typ        int
	Zone       int
	Index      int
}

// CreateAcceptor 创建监听节点
func CreateAcceptor(param NetNodeParam) iface.INetNode {
	node := socket.NewServerNode(param.ServerTyp, param.ServerName, param.Addr)
	node.(common.ProcessorRPCBundle).SetMessageProc(new(socket.TCPMessageProcessor))
	node.(common.ProcessorRPCBundle).SetHooker(new(ServerEventHook))
	msgHandle := GetMsgHandle(0)
	node.(common.ProcessorRPCBundle).SetMsgHandle(msgHandle)

	property := node.(common.ServerNodeProperty)
	property.SetServerTyp(param.Typ)
	property.SetZone(param.Zone)
	property.SetIndex(param.Index)

	node.Start()
	plugins.ETCDRegister(node)
	return node
}

func Init() {
	err := plugins.InitServiceDiscovery("127.0.0.1:2379")
	if err != nil {
		log.Println("InitServiceDiscovery err:", err)
		return
	}
}

func WaitExitSignal() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL)
	<-ch
}

func Stop(node iface.INetNode) {
	if node == nil {
		return
	}
	node.Stop()
}
