package tests

import (
	"fmt"
	"github.com/ljhe/scream/3rd/logrus"
	"github.com/ljhe/scream/core/config"
	"github.com/ljhe/scream/core/process"
	"testing"
)

func TestNewProcess(t *testing.T) {
	// 模拟配置文件
	conf := config.ScreamConfig{
		Process: config.Process{
			Id:   "proc-1",
			Host: "127.0.0.1",
			Node: []config.Node{
				{Name: "game", Addr: "0.0.0.0:2701", Typ: 2, Zone: 9999, Index: 1, Etcd: "127.0.0.1:2379"},
			},
		},
		Log: logrus.LogConfig{
			LogName:  "test",
			LogLevel: 6,
		},
	}

	p := process.NewProcess(conf.Process.Id, conf.Process.Host)
	fmt.Println(p.ID())
	fmt.Println(p.GetHost())
	p.Start()
	p.WaitExitSignal()
}
