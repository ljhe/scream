package tests

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/ljhe/scream/common/encryption"
	"github.com/ljhe/scream/common/socket"
	"github.com/ljhe/scream/common/util"
	"github.com/ljhe/scream/pbgo"
	"log"
	"math/rand"
	"net/url"
	"testing"
	"time"
)

func TestWSConnector(t *testing.T) {
	//for i := 0; i < 10; i++ {
	createConnector()
	//}
	//service.WaitExitSignal()
}

func createConnector() {
	// 1. 构造 WebSocket 连接的 URL
	u := url.URL{Scheme: "ws", Host: "localhost:9001", Path: "/ws"}
	log.Printf("connecting to %s", u.String())

	// 2. 连接到 WebSocket 服务端
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial error:", err)
	}
	defer func(c *websocket.Conn) {
		err := c.Close()
		if err != nil {
			log.Println("close error:", err)
		}
	}(c)

	// 3. 启动 goroutine 接收服务端消息
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read error:", err)
				return
			}
			log.Printf("recv: %s", message)
		}
	}()

	// 4. 循环发送消息给服务端
	rand.Seed(time.Now().UnixNano())
	random := util.RandomIntRange(1, 5)
	ticker := time.NewTicker(time.Duration(random) * time.Second)
	defer ticker.Stop()
	count := 0
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			// RSA加密
			data := &pbgo.CSSendMsgReq{
				Msg: c.LocalAddr().String() + "_" + fmt.Sprintf("%d", count),
			}
			msgData, msgInfo, _ := socket.EncodeMessage(data)
			encryptStr, _ := encryption.RSAEncrypt(msgData, encryption.RSAWSPublicKey)
			mb := &socket.MsgBase{
				MsgId:  msgInfo.ID,
				FlagId: 1,
			}
			buf := mb.MarshalBytes(encryptStr)
			err = c.WriteMessage(websocket.BinaryMessage, buf)
			count++
			if count >= 3 {
				err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				if err != nil {
					panic(err)
				}
				return
			}
		}
	}
}
