package websocket

import (
	"context"
	"github.com/gorilla/websocket"
	"github.com/ljhe/scream/3rd/logrus"
	"github.com/ljhe/scream/core/iface"
	"github.com/ljhe/scream/core/socket"
	"github.com/ljhe/scream/core/socket/sessions"
	"github.com/ljhe/scream/def"
	"github.com/ljhe/scream/utils"
	"net"
	"net/http"
	"syscall"
)

type webSocketAcceptor struct {
	socket.RuntimeTag // 运行状态
	socket.Option     // socket相关设置
	socket.Processor  // 事件处理相关
	socket.NodeProp
	socket.ContextSet
	iface.ISessionManager // 会话管理

	listener net.Listener // 保存端口
	upgrader *websocket.Upgrader
	server   *http.Server
}

func (ws *webSocketAcceptor) Start() iface.INetNode {
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
		logrus.Panicf("webSocketAcceptor listen fail. err:%v", err)
	}
	ws.listener = ln
	logrus.Infof("ws listen success ip:%v", ws.GetAddr())

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
			logrus.Errorf("ws listen field err:%v", err)
		}

		ws.SetRunState(false)
		ws.SetCloseFlag(false)
		ws.StopWg.Done()
	}()
	return ws
}

func (ws *webSocketAcceptor) Stop() {

}

func (ws *webSocketAcceptor) GetTyp() string {
	return def.SocketTypTcpWSAcceptor
}

// handleConnTest 测试websocket连接
func (ws *webSocketAcceptor) handleConnTest(w http.ResponseWriter, r *http.Request) {
	conn, err := ws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.Errorf("ws acceptor err:%v ip:%v", err, ws.GetAddr())
		return
	}

	ip, _ := utils.GetClientRealIP(r)
	// 读取消息
	for {
		typ, msg, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				logrus.Infof("ws client closed connection. remoteAddr:%v ip:%v", r.RemoteAddr, ip)
				return
			}
			logrus.Errorf("ws acceptor read message err:%v ip:%v", err, ws.GetAddr())
			return
		}

		// 打印客户端发送的数据
		logrus.Infof("ws acceptor receive msg:%v", string(msg))
		// 回复客户端
		if err := conn.WriteMessage(typ, msg); err != nil {
			return
		}
	}
}

func (ws *webSocketAcceptor) handleConn(w http.ResponseWriter, r *http.Request) {
	conn, err := ws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.Errorf("ws acceptor err:%v ip:%v", err, ws.GetAddr())
		return
	}

	sess := sessions.NewWSSession(conn, ws)
	sess.Start()
	// 通知上层事件(这边的回调要放到队列中，否则会有多线程冲突)
	ws.ProcEvent(&socket.RcvProcEvent{Sess: sess, Message: &socket.SessionAccepted{}})
}

func init() {
	socket.RegisterServerNode(func() iface.INetNode {
		node := &webSocketAcceptor{
			ISessionManager: sessions.NewSessionManager(),
			upgrader: &websocket.Upgrader{
				CheckOrigin: func(r *http.Request) bool {
					// 允许所有跨域请求 实际使用时需要谨慎设置
					return true
				},
			},
		}
		return node
	})
}
