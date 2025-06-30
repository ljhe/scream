package tests

import (
	"fmt"
	"github.com/ljhe/scream/3rd/logrus"
	"github.com/ljhe/scream/core/config"
	"github.com/ljhe/scream/core/iface"
	"github.com/ljhe/scream/core/process"
	"github.com/ljhe/scream/utils"
	"testing"
)

// 模拟配置文件
var conf = &config.ScreamConfig{
	Node: config.Node{
		Name:  "game",
		IP:    "127.0.0.1",
		Port:  2702,
		Typ:   2,
		Index: 1,
		Etcd:  "127.0.0.1:2379",
	},
	Log: logrus.LogConfig{
		LogName:  "game",
		LogLevel: 6,
	},
}

func TestNewProcess(t *testing.T) {
	p := &process.Process{
		P:     conf,
		Nodes: make([]iface.INetNode, 0),
	}

	p.Init()
	p.Start()
	p.WaitClose()
}

func TestDiscover(t *testing.T) {
	p := &process.Process{
		P:     conf,
		Nodes: make([]iface.INetNode, 0),
	}

	p.Init()
	p.Start()

	prop := p.Nodes[0].(iface.INodeProp)
	fmt.Println(p.Discover.GetNodeByKey(utils.ServerPreKey + utils.GenSelfServiceId(prop.GetName(), prop.GetServerTyp(), prop.GetIndex())))

	p.WaitClose()
}
