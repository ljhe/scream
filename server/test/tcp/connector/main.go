package main

import (
	"github.com/ljhe/scream/common"
	"github.com/ljhe/scream/common/config"
	"github.com/ljhe/scream/common/service"
	plugins "github.com/ljhe/scream/plugins/etcd"
	"github.com/ljhe/scream/plugins/logrus"
	"log"
)

func main() {
	*config.ServerConfigPath = "./server/test/tcp/connector/config.yaml"
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
	service.Stop(connector.GetNodeByName("tests"))
	log.Println("server close")
}
