package tests

import (
	"github.com/ljhe/scream/3rd/logrus"
	"github.com/ljhe/scream/core/config"
	"github.com/ljhe/scream/core/iface"
	"github.com/ljhe/scream/core/process"
	"testing"
)

func TestNewProcess(t *testing.T) {
	// 模拟配置文件
	conf := &config.ScreamConfig{
		Node: config.Node{
			Name:  "game",
			IP:    "127.0.0.1",
			Port:  2702,
			Typ:   2,
			Zone:  1,
			Index: 1,
			Etcd:  "127.0.0.1:2379",
		},
		Log: logrus.LogConfig{
			LogName:  "game",
			LogLevel: 6,
		},
	}

	p := &process.Process{
		P:     conf,
		Nodes: make([]iface.INetNode, 0),
	}

	p.Init()
	p.Start()
	p.WaitClose()
}
