package service

import (
	"github.com/ljhe/scream/common"
	_ "github.com/ljhe/scream/common/baseserver/normal_logic"
	"github.com/ljhe/scream/common/config"
	"github.com/ljhe/scream/common/iface"
	"github.com/ljhe/scream/common/socket"
	_ "github.com/ljhe/scream/common/socket/tcp"
	_ "github.com/ljhe/scream/common/socket/websocket"
	"github.com/ljhe/scream/pbgo"
	plugins "github.com/ljhe/scream/plugins/etcd"
	"time"
)

func GateWsFrontEndOpt() []iface.Option {
	var options []iface.Option
	options = append(options, func(s iface.INetNode) {
		bundle, ok := s.(common.ProcessorRPCBundle)
		if ok {
			bundle.SetMessageProc(new(socket.WSMessageProcessor)) //socket 收发数据处理
			bundle.(common.ProcessorRPCBundle).SetHooker(new(WsEventHook))
			msgPrcFunc := pbgo.GetMessageHandler(common.ServiceNodeTypeGateStr)
			bundle.(common.ProcessorRPCBundle).SetMsgRouter(msgPrcFunc)
		}
	})
	return options
}

// CreateAcceptor 创建监听节点
func CreateAcceptor() iface.INetNode {
	node := socket.NewServerNode(common.SocketTypTcpAcceptor, config.SConf.Node.Name, config.SConf.Node.Addr)
	node.(common.ProcessorRPCBundle).SetMessageProc(new(socket.TCPMessageProcessor))
	node.(common.ProcessorRPCBundle).SetHooker(new(ServerEventHook))
	msgHandle := GetMsgHandle(100)
	node.(common.ProcessorRPCBundle).SetMsgHandle(msgHandle)

	msgPrcFunc := pbgo.GetMessageHandler(common.ServiceNodeTypeGameStr)
	node.(common.ProcessorRPCBundle).SetMsgRouter(msgPrcFunc)

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
func CreateConnector(connect string, multiNode plugins.MultiServerNode) {
	plugins.DiscoveryService(multiNode, connect, config.SConf.Node.Zone,
		func(mn plugins.MultiServerNode, ed *plugins.ETCDServiceDesc) {
			// 不连接自己
			if ed.Typ == config.SConf.Node.Typ && ed.Zone == config.SConf.Node.Zone && ed.Index == config.SConf.Node.Index {
				return
			}
			node := socket.NewServerNode(common.SocketTypTcpConnector, config.SConf.Node.Name, ed.Host)
			msgHandle := GetMsgHandle(0)
			node.(common.ProcessorRPCBundle).SetHooker(new(ServerEventHook))
			node.(common.ProcessorRPCBundle).SetMsgHandle(msgHandle)
			node.(common.ProcessorRPCBundle).SetMessageProc(new(socket.TCPMessageProcessor))

			if opt, ok := node.(common.TCPSocketOption); ok {
				opt.SetSocketBuff(common.MsgMaxLen, common.MsgMaxLen, true)
				// 15s无读写断开 服务器之间已经添加心跳来维持读写
				opt.SetSocketDeadline(time.Second*15, time.Second*15)
			}

			property := node.(common.ServerNodeProperty)
			property.SetServerTyp(config.SConf.Node.Typ)
			property.SetZone(config.SConf.Node.Zone)
			property.SetIndex(config.SConf.Node.Index)

			// 将etcd信息保存在内存中
			node.(common.ContextSet).SetContextData(common.ContextSetEtcdKey, ed)
			// 添加到服务发现的节点管理中
			mn.AddNode(ed, node)

			node.Start()
		})
}

// CreateWebSocketAcceptor 创建监听节点
func CreateWebSocketAcceptor(opts ...iface.Option) iface.INetNode {
	node := socket.NewServerNode(common.SocketTypTcpWSAcceptor, config.SConf.Node.Name, config.SConf.Node.WsAddr)

	for _, opt := range opts {
		opt(node)
	}

	if opt, ok := node.(common.TCPSocketOption); ok {
		opt.SetSocketBuff(common.MsgMaxLen, common.MsgMaxLen, true)
		// 40秒无读 30秒无写断开 如果没有心跳了超时直接断开 调试期间可以不加
		// 通过该方法来模拟心跳保持连接
		opt.SetSocketDeadline(time.Second*40, time.Second*40)
		// 读/写协程没有过滤超时事件 发生了操时操作就断开连接
	}

	property := node.(common.ServerNodeProperty)
	property.SetServerTyp(config.SConf.Node.Typ)
	property.SetZone(config.SConf.Node.Zone)
	property.SetIndex(config.SConf.Node.Index)

	node.Start()

	// 注册到服务发现etcd中
	plugins.ETCDRegister(node)
	return node
}
