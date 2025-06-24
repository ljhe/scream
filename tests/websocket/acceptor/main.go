package main

import (
	"github.com/ljhe/scream/core/config"
	"github.com/ljhe/scream/core/process"
)

func main() {
	if *config.ServerConfigPath == "" {
		*config.ServerConfigPath = "./tests/websocket/acceptor/config.yaml"
	}
	p := process.NewProcess()
	p.Init()
	p.Start()
	p.WaitClose()
}
