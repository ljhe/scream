package main

import (
	"common"
	"common/plugins/logrus"
	"common/service"
)

func main() {
	logrus.Log(logrus.LogsSystem).Info("server starting ...")
	service.Init()
	node := service.CreateAcceptor(service.NetNodeParam{
		ServerTyp:  common.SocketTypTcpAcceptor,
		ServerName: "test",
		Addr:       "0.0.0.0:2701",
		Typ:        1,
		Zone:       9999,
		Index:      1,
	})
	logrus.Log(logrus.LogsSystem).Info("server start success")
	service.WaitExitSignal()
	logrus.Log(logrus.LogsSystem).Info("server stopping ...")
	service.Stop(node)
	logrus.Log(logrus.LogsSystem).Info("server close")
}
