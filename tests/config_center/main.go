package main

import (
	"github.com/ljhe/scream/3rd/logrus"
	"github.com/ljhe/scream/core/config"
	"github.com/ljhe/scream/core/service"
	"github.com/ljhe/scream/tests/config_center/manager"
)

func main() {
	*config.ServerConfigPath = "./tests/config_center/config.yaml"
	err := service.Init()
	if err != nil {
		logrus.Panicf("server starting fail:%v", err)
	}
	manager.NewCenter().Run()
	service.WaitExitSignal()
}
