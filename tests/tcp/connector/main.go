package main

import (
	"github.com/ljhe/scream/3rd/logrus"
	"github.com/ljhe/scream/core/config"
	"github.com/ljhe/scream/core/service"
)

func main() {
	*config.ServerConfigPath = "./tests/tcp/connector/config.yaml"
	err := service.Init()
	if err != nil {
		logrus.Panicf("server starting fail:%v", err)
	}
	service.StartUp()
}
