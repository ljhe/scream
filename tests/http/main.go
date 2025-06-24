package main

import (
	"github.com/ljhe/scream/3rd/logrus"
	"github.com/ljhe/scream/core/socket/http"
	"github.com/ljhe/scream/utils"
)

func main() {
	logrus.Init("")
	server := http.NewHttpServer()
	node := server.Start()
	utils.WaitExitSignal()
	node.Stop()
}
