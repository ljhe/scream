package service

import (
	trdetcd "github.com/ljhe/scream/3rd/etcd"
	_ "github.com/ljhe/scream/core/baseserver/normal_logic"
	"github.com/ljhe/scream/core/config"
	"github.com/ljhe/scream/core/iface"
	"github.com/ljhe/scream/core/socket"
	_ "github.com/ljhe/scream/core/socket/tcp"
	_ "github.com/ljhe/scream/core/socket/websocket"
	"github.com/ljhe/scream/def"
	"github.com/ljhe/scream/pbgo"
	"time"
)

// CreateAcceptor 创建监听节点
func CreateAcceptor() iface.INetNode {
	node := socket.NewServerNode(def.SocketTypTcpAcceptor, config.SConf.Node.Name, config.SConf.Node.Addr)
	node.(iface.IProcessor).SetHooker(new(socket.ServerHookEvent))
	node.(iface.IProcessor).SetMsgHandle(GetMsgHandle())
	node.(iface.IProcessor).SetMsgFlow(new(socket.TCPMsgFlow))

	if opt, ok := node.(iface.IOption); ok {
		// 15s无读写断开 服务器之间已经添加心跳来维持读写
		opt.SetOption(&socket.Option{
			ReadBufferSize:  def.MsgMaxLen,
			WriteBufferSize: def.MsgMaxLen,
			ReadTimeout:     time.Second * 15,
			WriteTimeout:    time.Second * 15,
		})
	}

	msgPrcFunc := pbgo.GetMessageHandler(def.GetServiceNodeStr(config.SConf.Node.Typ))
	node.(iface.IProcessor).SetMsgRouter(msgPrcFunc)

	node.(iface.INodeProp).SetNodeProp()

	node.Start()

	// 注册到服务发现etcd中
	trdetcd.Register(node)
	// 加载数据到discover
	node.(iface.IDiscover).Loader()
	return node
}

// CreateConnector 创建连接节点
func CreateConnector(connect string) {
	trdetcd.DiscoveryService(connect, config.SConf.Node.Zone, func(ed *trdetcd.ServerInfo) {
		// 不连接自己
		if ed.Typ == config.SConf.Node.Typ && ed.Zone == config.SConf.Node.Zone && ed.Index == config.SConf.Node.Index {
			return
		}
		node := socket.NewServerNode(def.SocketTypTcpConnector, config.SConf.Node.Name, ed.Host)
		node.(iface.IProcessor).SetHooker(new(socket.ServerHookEvent))
		node.(iface.IProcessor).SetMsgHandle(GetMsgHandle())
		node.(iface.IProcessor).SetMsgFlow(new(socket.TCPMsgFlow))

		if opt, ok := node.(iface.IOption); ok {
			// 15s无读写断开 服务器之间已经添加心跳来维持读写
			opt.SetOption(&socket.Option{
				ReadBufferSize:  def.MsgMaxLen,
				WriteBufferSize: def.MsgMaxLen,
				ReadTimeout:     time.Second * 15,
				WriteTimeout:    time.Second * 15,
			})
		}

		node.(iface.INodeProp).SetNodeProp()

		// 将etcd信息保存在内存中
		node.(iface.IContextSet).SetContextData(def.ContextSetEtcdKey, ed)

		node.Start()
	})
}

// CreateWebSocketAcceptor 创建监听节点
func CreateWebSocketAcceptor() iface.INetNode {
	node := socket.NewServerNode(def.SocketTypTcpWSAcceptor, config.SConf.Node.Name, config.SConf.Node.WsAddr)

	node.(iface.IProcessor).SetMsgFlow(new(socket.WSMsgFlow))
	node.(iface.IProcessor).SetHooker(new(socket.WsHookEvent))
	msgPrcFunc := pbgo.GetMessageHandler(def.GetServiceNodeStr(config.SConf.Node.Typ))
	node.(iface.IProcessor).SetMsgRouter(msgPrcFunc)

	if opt, ok := node.(iface.IOption); ok {
		// 40秒无读 30秒无写断开 如果没有心跳了超时直接断开 调试期间可以不加
		// 通过该方法来模拟心跳保持连接
		// 读/写协程没有过滤超时事件 发生了操时操作就断开连接
		opt.SetOption(&socket.Option{
			ReadBufferSize:  def.MsgMaxLen,
			WriteBufferSize: def.MsgMaxLen,
			ReadTimeout:     time.Second * 40,
			WriteTimeout:    time.Second * 30,
		})
	}

	node.(iface.INodeProp).SetNodeProp()

	node.Start()
	return node
}
