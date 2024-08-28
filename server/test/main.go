package main

import (
	"common"
	plugins "common/plugins/etcd"
	"common/service"
	"log"
)

func main() {
	log.Println("server starting ...")
	service.Init()
	connector := plugins.NewMultiServerNode()
	service.CreateConnector(service.NetNodeParam{
		ServerTyp:            common.SocketTypTcpConnector,
		ServerName:           "test_connector",
		Typ:                  2,
		Zone:                 9999,
		Index:                1,
		DiscoveryServiceName: "test",
	}, connector)
	log.Println("server start success")
	service.WaitExitSignal()
	log.Println("server stopping ...")
	service.Stop(connector.GetNodeByName("test"))
	log.Println("server close")
}
