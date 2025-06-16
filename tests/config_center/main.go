package main

import (
	"github.com/ljhe/scream/3rd/logrus"
	"github.com/ljhe/scream/core/service"
	"github.com/ljhe/scream/tests/config_center/manager"
)

func main() {
	logrus.Init("")
	c := manager.NewCenter()
	c.Run()
	service.WaitExitSignal()
}
