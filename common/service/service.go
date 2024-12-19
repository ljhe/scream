package service

import (
	"common"
	"common/config"
	"common/iface"
	plugins "common/plugins/etcd"
	"common/plugins/logrus"
	"common/plugins/mpool"
	"common/socket"
	_ "common/socket/tcp"
	"os"
	"os/signal"
	"syscall"
)

type NetNodeParam struct {
	ServerTyp            string
	ServerName           string
	Addr                 string
	Typ                  int
	Zone                 int
	Index                int
	DiscoveryServiceName string // 用于服务发现
}

// CreateAcceptor 创建监听节点
func CreateAcceptor(serverTyp string) iface.INetNode {
	node := socket.NewServerNode(serverTyp, config.SConf.Node.Name, config.SConf.Node.Addr)
	node.(common.ProcessorRPCBundle).SetMessageProc(new(socket.TCPMessageProcessor))
	node.(common.ProcessorRPCBundle).SetHooker(new(ServerEventHook))
	msgHandle := GetMsgHandle(0)
	node.(common.ProcessorRPCBundle).SetMsgHandle(msgHandle)

	property := node.(common.ServerNodeProperty)
	property.SetServerTyp(config.SConf.Node.Typ)
	property.SetZone(config.SConf.Node.Zone)
	property.SetIndex(config.SConf.Node.Index)

	node.Start()

	// 注册到服务发现etcd中
	plugins.ETCDRegister(node)
	return node
}

// CreateConnector 创建连接节点
func CreateConnector(serverTyp string, multiNode plugins.MultiServerNode) iface.INetNode {
	plugins.DiscoveryService(multiNode, config.SConf.Node.DiscoveryServiceName, config.SConf.Node.Zone,
		func(mn plugins.MultiServerNode, ed *plugins.ETCDServiceDesc) {
			// 不连接自己
			if ed.Typ == config.SConf.Node.Typ && ed.Zone == config.SConf.Node.Zone && ed.Index == config.SConf.Node.Index {
				return
			}
			node := socket.NewServerNode(serverTyp, config.SConf.Node.Name, ed.Host)
			msgHandle := GetMsgHandle(0)
			node.(common.ProcessorRPCBundle).SetHooker(new(ServerEventHook))
			node.(common.ProcessorRPCBundle).SetMsgHandle(msgHandle)
			node.(common.ProcessorRPCBundle).SetMessageProc(new(socket.TCPMessageProcessor))

			property := node.(common.ServerNodeProperty)
			property.SetServerTyp(config.SConf.Node.Typ)
			property.SetZone(config.SConf.Node.Zone)
			property.SetIndex(config.SConf.Node.Index)

			// 将etcd信息保存在内存中
			node.(common.ContextSet).SetContextData(common.ContextSetEtcdKey, ed)
			// 添加到服务发现的节点管理中
			mn.AddNode(config.SConf.Node.DiscoveryServiceName, ed, node)

			node.Start()
		})
	return nil
}

func Init() error {
	// 加载系统配置文件
	config.Init()
	// 初始化日志模块
	logrus.Init(*config.ServerConfigPath)
	// 初始化内存池
	mpool.MemoryPoolInit()
	// 初始化服务发现
	err := plugins.InitServiceDiscovery("127.0.0.1:2379")
	if err != nil {
		logrus.Log(logrus.LogsSystem).Errorf("InitServiceDiscovery err:%v", err)
		return err
	}
	return nil
}

func WaitExitSignal() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL)
	<-ch
}

func Stop(node iface.INetNode) {
	if node == nil {
		return
	}
	node.Stop()
}
