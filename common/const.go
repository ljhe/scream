package common

const (
	SocketTypTcpAcceptor   = "tcpAcceptor"
	SocketTypTcpConnector  = "tcpConnector"
	SocketTypTcpWSAcceptor = "tcpWebSocketAcceptor"
)

const (
	ContextSetEtcdKey = "etcd_node"
	ContextSetCtxKey  = "ctx"
)

const MsgMaxLen = 1024 * 40 // 40k(发送和接受字节最大数量)

const (
	MsgEncryptionRSA = 1
)
