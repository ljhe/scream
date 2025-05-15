package websocket

import (
	"context"
	"github.com/gorilla/websocket"
	"github.com/ljhe/scream/common"
	"github.com/ljhe/scream/common/iface"
	"github.com/ljhe/scream/common/socket"
	"github.com/ljhe/scream/common/util"
	"github.com/ljhe/scream/plugins/logrus"
	"log"
	"net"
	"net/http"
	"syscall"
)

type tcpWebSocketAcceptor struct {
	socket.RuntimeTag      // 运行状态
	socket.TCPSocketOption // socket相关设置
	socket.Processor       // 事件处理相关
	socket.ServerNodeProperty
	socket.ContextSet
	iface.ISessionManager // 会话管理

	listener net.Listener // 保存端口
	upgrader *websocket.Upgrader
	server   *http.Server
}

func (ws *tcpWebSocketAcceptor) Start() iface.INetNode {
	// 正在停止的话 需要先等待
	ws.StopWg.Wait()
	// 防止重入导致错误
	if ws.GetRunState() {
		return ws
	}

	var listenCfg = net.ListenConfig{
		Control: func(network, address string, c syscall.RawConn) error {
			var controlErr error
			err := c.Control(func(fd uintptr) {
				controlErr = syscall.SetsockoptInt(syscall.Handle(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
				return
			})
			if err != nil {
				return err
			}
			return controlErr
		},
	}
	ln, err := listenCfg.Listen(context.Background(), "tcp", ws.GetAddr())
	if err != nil {
		log.Panicf("webSocketAcceptor listen fail. err:%v", err)
	}
	ws.listener = ln
	logrus.Log(logrus.LogsSystem).Infof("ws listen success ip:%v", ws.GetAddr())

	// 是否正在结束中
	if ws.GetCloseFlag() {
		return ws
	}
	ws.SetRunState(true)

	mux := http.NewServeMux()
	mux.HandleFunc("/", ws.handleConn)

	ws.server = &http.Server{Addr: ws.GetAddr(), Handler: mux}
	go func() {
		err = ws.server.Serve(ws.listener)
		if err != nil {
			logrus.Log(logrus.LogsSystem).Errorf("ws listen field err:%v", err)
		}

		ws.SetRunState(false)
		ws.SetCloseFlag(false)
		ws.StopWg.Done()
	}()
	return ws
}

func (ws *tcpWebSocketAcceptor) Stop() {

}

func (ws *tcpWebSocketAcceptor) GetTyp() string {
	return common.SocketTypTcpWSAcceptor
}

// handleConnTest 测试websocket连接
func (ws *tcpWebSocketAcceptor) handleConnTest(w http.ResponseWriter, r *http.Request) {
	conn, err := ws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.Log(logrus.LogsSystem).Errorf("ws acceptor err:%v ip:%v", err, ws.GetAddr())
		return
	}

	ip, _ := util.GetClientRealIP(r)
	// 读取消息
	for {
		typ, msg, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				logrus.Log(logrus.LogsSystem).Infof("ws client closed connection. remoteAddr:%v ip:%v", r.RemoteAddr, ip)
				return
			}
			logrus.Log(logrus.LogsSystem).Errorf("ws acceptor read message err:%v ip:%v", err, ws.GetAddr())
			return
		}

		// 打印客户端发送的数据
		logrus.Log(logrus.LogsSystem).Infof("ws acceptor receive msg:%v", string(msg))
		// 回复客户端
		if err := conn.WriteMessage(typ, msg); err != nil {
			return
		}
	}
}

func (ws *tcpWebSocketAcceptor) handleConn(w http.ResponseWriter, r *http.Request) {
	conn, err := ws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.Log(logrus.LogsSystem).Errorf("ws acceptor err:%v ip:%v", err, ws.GetAddr())
		return
	}

	ws.SocketOptWebSocket(conn)
	sess := newWSSession(conn, ws, nil)
	sess.start()
	// 通知上层事件(这边的回调要放到队列中，否则会有多线程冲突)
	ws.ProcEvent(&socket.RcvMsgEvent{Sess: sess, Message: &socket.SessionAccepted{}})
}

func init() {
	socket.RegisterServerNode(func() iface.INetNode {
		node := &tcpWebSocketAcceptor{
			ISessionManager: socket.NewSessionManager(),
			upgrader: &websocket.Upgrader{
				CheckOrigin: func(r *http.Request) bool {
					// 允许所有跨域请求 实际使用时需要谨慎设置
					return true
				},
			},
		}
		node.TCPSocketOption.Init()
		return node
	})
	log.Println("ws acceptor register success.")
}
