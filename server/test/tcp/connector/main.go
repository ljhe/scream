package main

import (
	"common"
	"common/config"
	plugins "common/plugins/etcd"
	"common/plugins/logrus"
	"common/service"
	"log"
)

func main() {
	*config.ServerConfigPath = "./test/tcp/connector/config.yaml"
	err := service.Init()
	if err != nil {
		logrus.Log(logrus.LogsSystem).Errorf("server starting fail:%v", err)
		return
	}
	logrus.Log(logrus.LogsSystem).Info("server starting ...")
	connector := plugins.NewMultiServerNode()
	service.CreateConnector(common.SocketTypTcpConnector, connector)
	log.Println("server start success")
	service.WaitExitSignal()
	log.Println("server stopping ...")
	service.Stop(connector.GetNodeByName("test"))
	log.Println("server close")
}
