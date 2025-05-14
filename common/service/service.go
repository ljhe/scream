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
	node.(iface.IProcessor).SetMessageProc(new(socket.TCPMessageProcessor))
	node.(iface.IProcessor).SetHooker(new(socket.ServerHookEvent))
	msgHandle := GetMsgHandle(100)
	node.(iface.IProcessor).SetMsgHandle(msgHandle)

	msgPrcFunc := pbgo.GetMessageHandler(common.GetServiceNodeStr(config.SConf.Node.Typ))
	node.(iface.IProcessor).SetMsgRouter(msgPrcFunc)

	node.(iface.IServerNodeProperty).SetServerNodeProperty()

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
			node.(iface.IProcessor).SetHooker(new(socket.ServerHookEvent))
			node.(iface.IProcessor).SetMsgHandle(msgHandle)
			node.(iface.IProcessor).SetMessageProc(new(socket.TCPMessageProcessor))

			if opt, ok := node.(iface.ITCPSocketOption); ok {
				opt.SetSocketBuff(common.MsgMaxLen, common.MsgMaxLen, true)
				// 15s无读写断开 服务器之间已经添加心跳来维持读写
				opt.SetSocketDeadline(time.Second*15, time.Second*15)
			}

			node.(iface.IServerNodeProperty).SetServerNodeProperty()

			// 将etcd信息保存在内存中
			node.(iface.IContextSet).SetContextData(common.ContextSetEtcdKey, ed)
			// 添加到服务发现的节点管理中
			mn.AddNode(ed, node)

			node.Start()
		})
}

// CreateWebSocketAcceptor 创建监听节点
func CreateWebSocketAcceptor() iface.INetNode {
	node := socket.NewServerNode(common.SocketTypTcpWSAcceptor, config.SConf.Node.Name, config.SConf.Node.WsAddr)

	//node.(iface.INetProcessor).SetMessageProc(new(socket.WSMessageProcessor))
	//node.(iface.INetProcessor).(iface.INetProcessor).SetHooker(new(WsHookEvent))
	msgPrcFunc := pbgo.GetMessageHandler(common.GetServiceNodeStr(config.SConf.Node.Typ))
	node.(iface.IProcessor).(iface.IProcessor).SetMsgRouter(msgPrcFunc)

	if opt, ok := node.(iface.ITCPSocketOption); ok {
		opt.SetSocketBuff(common.MsgMaxLen, common.MsgMaxLen, true)
		// 40秒无读 30秒无写断开 如果没有心跳了超时直接断开 调试期间可以不加
		// 通过该方法来模拟心跳保持连接
		opt.SetSocketDeadline(time.Second*40, time.Second*30)
		// 读/写协程没有过滤超时事件 发生了操时操作就断开连接
	}

	node.(iface.IServerNodeProperty).SetServerNodeProperty()

	node.Start()

	// 注册到服务发现etcd中
	plugins.ETCDRegister(node)
	return node
}
