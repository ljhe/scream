package process

import (
	"fmt"
	"github.com/ljhe/scream/3rd/logrus"
	"github.com/ljhe/scream/core/iface"
	"github.com/ljhe/scream/core/system"
	"github.com/ljhe/scream/utils"
)

type Process struct {
	p   param
	sys iface.ISystem
}

type param struct {
	ID     string // node's globally unique ID
	Weight int

	Ip   string
	Port int

	Loader  iface.INodeLoader
	Factory iface.INodeFactory
}

type Option func(*param)

func WithServiceInfo(ip string, port int) Option {
	return func(p *param) {
		p.Ip = ip
		p.Port = port
	}
}

func WithID(id string) Option {
	return func(np *param) {
		np.ID = id
	}
}

func WithWeight(weight int) Option {
	return func(np *param) {
		np.Weight = weight
	}
}

func WithLoader(load iface.INodeLoader) Option {
	return func(p *param) {
		p.Loader = load
	}
}

func WithFactory(factory iface.INodeFactory) Option {
	return func(p *param) {
		p.Factory = factory
	}
}

func WithIP(ip string) Option {
	return func(np *param) {
		np.Ip = ip
	}
}

func WithPort(port int) Option {
	return func(np *param) {
		np.Port = port
	}
}

var pcs *Process

func BuildProcessWithOption(opts ...Option) iface.IProcess {
	p := param{
		Ip: "127.0.0.1",
	}

	for _, opt := range opts {
		opt(&p)
	}

	pcs = &Process{
		sys: system.BuildSystemWithOption(p.ID, p.Ip, p.Port, p.Loader, p.Factory),
		p:   p,
	}
	return pcs
}

func Get() iface.IProcess {
	return pcs
}

func (p *Process) Init() error {
	// 加载系统配置文件
	//if !utils.IsTesting() {
	//	p.p = config.Init()
	//}
	//// 初始化日志模块
	//logrus.Init(*config.ServerConfigPath)
	//// 初始化服务发现
	//err := trdetcd.InitServiceDiscovery(p.P.Node.Etcd)
	//if err != nil {
	//	logrus.Errorf("InitServiceDiscovery err:%v", err)
	//	return err
	//}
	//// 初始化db
	//config.OrmConnector = gorm.NewOrmConn()
	//err = config.OrmConnector.Start("root:123456@(127.0.0.1:3306)/gamedb_9999?charset=utf8&loc=Asia%2FShanghai&parseTime=true")
	//if err != nil {
	//	logrus.Errorf("init db err:%v", err)
	//	return err
	//}
	p.p.Loader.AssignToNode(p)
	return nil
}

func (p *Process) Start() error {
	logrus.Infof(fmt.Sprintf("[ %s ] starting ...", p.p.ID))

	logrus.Infof(fmt.Sprintf("[ %s ] started SUCCESS. ip:%s port:%d", p.p.ID, p.p.Ip, p.p.Port))
	return nil
}

func (p *Process) WaitClose() error {
	utils.WaitExitSignal()

	logrus.Infof(fmt.Sprintf("[ %s ] stoping ...", p.p.ID))
	return p.Stop()
}

func (p *Process) Stop() error {

	logrus.Infof(fmt.Sprintf("[ %s ] close ...", p.p.ID))
	return nil
}

func (p *Process) ID() string {
	return p.p.ID
}

func (p *Process) System() iface.ISystem {
	return p.sys
}
