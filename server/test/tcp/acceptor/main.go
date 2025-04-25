package main

import (
	"common"
	"common/config"
	"common/service"
	"plugins/logrus"
)

func main() {
	*config.ServerConfigPath = "./test/tcp/acceptor/config.yaml"
	err := service.Init()
	if err != nil {
		logrus.Log(logrus.LogsSystem).Errorf("server starting fail:%v", err)
		return
	}
	logrus.Log(logrus.LogsSystem).Info("server starting ...")
	node := service.CreateAcceptor(common.SocketTypTcpAcceptor)
	logrus.Log(logrus.LogsSystem).Info("server start success")
	service.WaitExitSignal()
	logrus.Log(logrus.LogsSystem).Info("server stopping ...")
	service.Stop(node)
	logrus.Log(logrus.LogsSystem).Info("server close")
}
