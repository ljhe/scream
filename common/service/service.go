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

// CreateAcceptor 创建监听节点
func CreateAcceptor() iface.INetNode {
	node := socket.NewServerNode(common.SocketTypTcpAcceptor, config.SConf.Node.Name, config.SConf.Node.Addr)
	node.(iface.ProcessorRPCBundle).SetMessageProc(new(socket.TCPMessageProcessor))
	node.(iface.ProcessorRPCBundle).SetHooker(new(ServerHookEvent))
	msgHandle := GetMsgHandle(100)
	node.(iface.ProcessorRPCBundle).SetMsgHandle(msgHandle)

	msgPrcFunc := pbgo.GetMessageHandler(common.ServiceNodeTypeGameStr)
	node.(iface.ProcessorRPCBundle).SetMsgRouter(msgPrcFunc)

	node.(iface.ServerNodeProperty).SetServerNodeProperty()

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
			node.(iface.ProcessorRPCBundle).SetHooker(new(ServerHookEvent))
			node.(iface.ProcessorRPCBundle).SetMsgHandle(msgHandle)
			node.(iface.ProcessorRPCBundle).SetMessageProc(new(socket.TCPMessageProcessor))

			if opt, ok := node.(iface.TCPSocketOption); ok {
				opt.SetSocketBuff(common.MsgMaxLen, common.MsgMaxLen, true)
				// 15s无读写断开 服务器之间已经添加心跳来维持读写
				opt.SetSocketDeadline(time.Second*15, time.Second*15)
			}

			node.(iface.ServerNodeProperty).SetServerNodeProperty()

			// 将etcd信息保存在内存中
			node.(iface.ContextSet).SetContextData(common.ContextSetEtcdKey, ed)
			// 添加到服务发现的节点管理中
			mn.AddNode(ed, node)

			node.Start()
		})
}

// CreateWebSocketAcceptor 创建监听节点
func CreateWebSocketAcceptor() iface.INetNode {
	node := socket.NewServerNode(common.SocketTypTcpWSAcceptor, config.SConf.Node.Name, config.SConf.Node.WsAddr)

	node.(iface.ProcessorRPCBundle).SetMessageProc(new(socket.WSMessageProcessor)) //socket 收发数据处理
	node.(iface.ProcessorRPCBundle).(iface.ProcessorRPCBundle).SetHooker(new(WsHookEvent))
	msgPrcFunc := pbgo.GetMessageHandler(common.ServiceNodeTypeGateStr)
	node.(iface.ProcessorRPCBundle).(iface.ProcessorRPCBundle).SetMsgRouter(msgPrcFunc)

	if opt, ok := node.(iface.TCPSocketOption); ok {
		opt.SetSocketBuff(common.MsgMaxLen, common.MsgMaxLen, true)
		// 40秒无读 30秒无写断开 如果没有心跳了超时直接断开 调试期间可以不加
		// 通过该方法来模拟心跳保持连接
		opt.SetSocketDeadline(time.Second*40, time.Second*40)
		// 读/写协程没有过滤超时事件 发生了操时操作就断开连接
	}

	node.(iface.ServerNodeProperty).SetServerNodeProperty()

	node.Start()

	// 注册到服务发现etcd中
	plugins.ETCDRegister(node)
	return node
}
