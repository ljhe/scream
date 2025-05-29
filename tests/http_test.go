package tests

import (
	"github.com/ljhe/scream/3rd/logrus"
	"github.com/ljhe/scream/core/service"
	"github.com/ljhe/scream/core/socket/http"
	"testing"
)

func TestAcceptor(t *testing.T) {
	logrus.Init("../3rd/logrus/config.yaml")
	go http.Server.Start()
	service.WaitExitSignal()
}
