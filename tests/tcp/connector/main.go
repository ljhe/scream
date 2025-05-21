package main

import (
	"github.com/ljhe/scream/core/config"
	"github.com/ljhe/scream/core/service"
)

func main() {
	*config.ServerConfigPath = "./tests/tcp/connector/config.yaml"
	service.StartUp()
}
