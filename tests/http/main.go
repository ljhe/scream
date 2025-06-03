package main

import (
	"github.com/ljhe/scream/3rd/logrus"
	"github.com/ljhe/scream/core/service"
	"github.com/ljhe/scream/core/socket/http"
)

func main() {
	logrus.Init("./3rd/logrus/config.yaml")
	server := http.NewHttpServer()
	node := server.Start()
	service.WaitExitSignal()
	node.Stop()
}
