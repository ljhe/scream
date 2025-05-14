package socket

import (
	"github.com/ljhe/scream/common/iface"
	"log"
)

type nodeCreate func() iface.INetNode

var serverNodeByTyp = map[string]nodeCreate{}

func RegisterServerNode(f nodeCreate) {
	node := f()
	if _, ok := serverNodeByTyp[node.GetTyp()]; ok {
		return
	}
	serverNodeByTyp[node.GetTyp()] = f
}

func NewServerNode(serverTyp, serverName, addr string) iface.INetNode {
	f := serverNodeByTyp[serverTyp]
	if f == nil {
		log.Printf("f is nil. typ:%s \n", serverTyp)
		return nil
	}
	node := f()
	nodeProperty := node.(iface.IServerNodeProperty)
	nodeProperty.SetAddr(addr)
	nodeProperty.SetName(serverName)
	return node
}
