package main

import (
	"common"
	"common/config"
	"common/iface"
	"common/plugins/logrus"
	"common/service"
	"common/socket"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

func GateWsFrontEndOpt() []iface.Option {
	var options []iface.Option
	options = append(options, func(s iface.INetNode) {
		bundle, ok := s.(common.ProcessorRPCBundle)
		if ok {
			bundle.SetMessageProc(new(socket.WSMessageProcessor)) //socket 收发数据处理
			bundle.(common.ProcessorRPCBundle).SetHooker(new(service.ServerEventHook))
			msgHandle := service.GetMsgHandle(0)
			bundle.(common.ProcessorRPCBundle).SetMsgHandle(msgHandle)
		}
	})
	return options
}

func main() {

	//wsTest()

	*config.ServerConfigPath = "./test/websocket/acceptor/config.yaml"
	err := service.Init()
	if err != nil {
		logrus.Log(logrus.LogsSystem).Errorf("server starting fail:%v", err)
		return
	}
	logrus.Log(logrus.LogsSystem).Info("server starting ...")
	node := service.CreateWebSocketAcceptor(common.SocketTypTcpWSAcceptor, GateWsFrontEndOpt()...)
	logrus.Log(logrus.LogsSystem).Info("server start success")
	service.WaitExitSignal()
	logrus.Log(logrus.LogsSystem).Info("server stopping ...")
	service.Stop(node)
	logrus.Log(logrus.LogsSystem).Info("server close")
}

func wsTest() {
	fmt.Println("Starting WebSocket server on :3101")
	http.HandleFunc("/ws", handleConnections)
	log.Fatal(http.ListenAndServe(":3101", nil))
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true // 允许所有跨域请求，实际使用时需要谨慎设置
		},
	}

	// 升级 HTTP 连接到 WebSocket 连接
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal("Failed to upgrade connection:", err)
		return
	}

	defer conn.Close()

	for {
		// 读取消息
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				log.Println("Client closed connection.")
				return
			}
			log.Printf("Error reading message: %v\n", err)
			return
		}
		//打印客户端发送的数据
		fmt.Printf("Message Received: %s\n", string(p))

		// 回复客户端
		if err := conn.WriteMessage(messageType, p); err != nil {
			log.Printf("Error writing message: %v\n", err)
			return
		}
	}
}
