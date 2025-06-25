package process

import (
	"fmt"
	"github.com/ljhe/scream/3rd/db/gorm"
	trdetcd "github.com/ljhe/scream/3rd/etcd"
	"github.com/ljhe/scream/3rd/logrus"
	"github.com/ljhe/scream/core/config"
	"github.com/ljhe/scream/core/iface"
	"github.com/ljhe/scream/utils"
)

type Process struct {
	P        *config.ScreamConfig
	Nodes    []iface.INetNode
	Discover iface.IDiscover
}

func NewProcess() iface.IProcess {
	return &Process{
		Nodes: make([]iface.INetNode, 0),
	}
}

func (p *Process) Init() error {
	// 加载系统配置文件
	if !utils.IsTesting() {
		p.P = config.Init()
	}
	// 初始化日志模块
	logrus.Init(*config.ServerConfigPath)
	// 初始化内存池
	//mpool.MemoryPoolInit()
	// 初始化服务发现
	err := trdetcd.InitServiceDiscovery(p.P.Node.Etcd)
	if err != nil {
		logrus.Errorf("InitServiceDiscovery err:%v", err)
		return err
	}
	// 初始化db
	config.OrmConnector = gorm.NewOrmConn()
	err = config.OrmConnector.Start("root:123456@(127.0.0.1:3306)/gamedb_9999?charset=utf8&loc=Asia%2FShanghai&parseTime=true")
	if err != nil {
		logrus.Errorf("init db err:%v", err)
		return err
	}
	return nil
}

func (p *Process) Start() error {
	logrus.Infof(fmt.Sprintf("[ %s ] starting ...", p.P.Node.Name))
	p.Nodes = append(p.Nodes, p.CreateAcceptor())

	if p.P.Node.WsAddr != "" {
		p.Nodes = append(p.Nodes, p.CreateWebSocketAcceptor())
	}

	for _, connect := range p.P.Node.Connect {
		p.CreateConnector(connect)
	}

	// 加载数据到discover
	p.Discover = NewDiscover()

	logrus.Infof(fmt.Sprintf("[ %s ] start success ...", p.P.Node.Name))
	return nil
}

func (p *Process) WaitClose() error {
	utils.WaitExitSignal()

	logrus.Infof(fmt.Sprintf("[ %s ] stoping ...", p.P.Node.Name))
	return p.Stop()
}

func (p *Process) Stop() error {
	p.Discover.Close()
	for _, node := range p.Nodes {
		if node == nil {
			continue
		}
		node.Stop()
		trdetcd.UnRegister(node)
	}
	logrus.Infof(fmt.Sprintf("[ %s ] close ...", p.P.Node.Name))
	return nil
}
