package service

import (
	"fmt"
	"github.com/ljhe/scream/common/config"
	"github.com/ljhe/scream/common/iface"
	"github.com/ljhe/scream/plugins/db/gorm"
	plugins "github.com/ljhe/scream/plugins/etcd"
	"github.com/ljhe/scream/plugins/logrus"
	"os"
	"os/signal"
	"syscall"
)

func Init() error {
	// 加载系统配置文件
	config.Init()
	// 初始化日志模块
	logrus.Init(*config.ServerConfigPath)
	// 初始化内存池
	//mpool.MemoryPoolInit()
	// 初始化服务发现
	err := plugins.InitServiceDiscovery("127.0.0.1:2379")
	if err != nil {
		logrus.Log(logrus.LogsSystem).Errorf("InitServiceDiscovery err:%v", err)
		return err
	}
	// 初始化db
	config.OrmConnector = gorm.NewOrmConn()
	err = config.OrmConnector.Start("root:123456@(127.0.0.1:3306)/gamedb_9999?charset=utf8&loc=Asia%2FShanghai&parseTime=true")
	if err != nil {
		logrus.Log(logrus.LogsSystem).Errorf("init db err:%v", err)
		return err
	}
	return nil
}

func StartUp() {
	err := Init()
	if err != nil {
		logrus.Log(logrus.LogsSystem).Errorf("server starting fail:%v", err)
		return
	}

	logrus.Log(logrus.LogsSystem).Info(fmt.Sprintf("[ %s ] starting ...", config.SConf.Node.Name))
	nodes := make([]iface.INetNode, 0)
	if config.SConf.Node.Addr != "" {
		nodes = append(nodes, CreateAcceptor())
	}

	if config.SConf.Node.WsAddr != "" {
		nodes = append(nodes, CreateWebSocketAcceptor(GateWsFrontEndOpt()...))
	}

	for _, connect := range config.SConf.Node.Connect {
		multiNode := plugins.NewMultiServerNode()
		CreateConnector(connect, multiNode)
	}
	logrus.Log(logrus.LogsSystem).Info(fmt.Sprintf("[ %s ] start success ...", config.SConf.Node.Name))

	WaitExitSignal()

	logrus.Log(logrus.LogsSystem).Info(fmt.Sprintf("[ %s ] stoping ...", config.SConf.Node.Name))
	for _, node := range nodes {
		Stop(node)
	}

	logrus.Log(logrus.LogsSystem).Info(fmt.Sprintf("[ %s ] close ...", config.SConf.Node.Name))
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
