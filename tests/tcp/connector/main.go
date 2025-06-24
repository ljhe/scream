package main

import (
	"github.com/ljhe/scream/core/config"
	"github.com/ljhe/scream/core/process"
)

func main() {
	*config.ServerConfigPath = "./tests/tcp/connector/config.yaml"
	p := process.NewProcess()
	p.Init()
	p.Start()
	p.WaitClose()
}
