package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/ljhe/scream/common/config"
	"github.com/ljhe/scream/common/service"
	"log"
	"net/http"
)

func main() {

	//wsTest()

	*config.ServerConfigPath = "./tests/websocket/acceptor/config.yaml"
	service.StartUp()
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
