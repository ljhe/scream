package main

import (
	"github.com/ljhe/scream/common/config"
	"github.com/ljhe/scream/common/service"
)

func main() {
	*config.ServerConfigPath = "./tests/tcp/connector/config.yaml"
	service.StartUp()
}
