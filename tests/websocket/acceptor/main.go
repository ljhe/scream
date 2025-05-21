package main

import (
	"github.com/ljhe/scream/core/config"
	"github.com/ljhe/scream/core/service"
)

func main() {
	*config.ServerConfigPath = "./tests/websocket/acceptor/config.yaml"
	service.StartUp()
}
