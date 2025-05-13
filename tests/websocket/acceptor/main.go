package main

import (
	"github.com/ljhe/scream/common/config"
	"github.com/ljhe/scream/common/service"
)

func main() {
	*config.ServerConfigPath = "./tests/websocket/acceptor/config.yaml"
	service.StartUp()
}
