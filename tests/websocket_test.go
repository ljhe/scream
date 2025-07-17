package tests

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/ljhe/scream/message"
	"github.com/ljhe/scream/pbgo"
	"github.com/ljhe/scream/utils"
	"log"
	"math/rand"
	"net/url"
	"sync"
	"testing"
	"time"
)

var wg sync.WaitGroup

func TestWSConnector(t *testing.T) {
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go createConnector()
	}
	wg.Wait()
}

func TestWSConnectorWithoutPing(t *testing.T) {
	c := connect()
	defer c.Close()

	time.Sleep(10 * time.Second)
	send(c, 1)
	utils.WaitExitSignal()
}

func createConnector() {
	c := connect()
	defer c.Close()

	stopReading := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			select {
			case <-stopReading:
				log.Println("read goroutine exiting. addr:", c.LocalAddr().String())
				return
			default:
				_, message, err := c.ReadMessage()
				if err != nil {
					log.Println("read error:", err)
					return
				}
				log.Printf("recv: %s", message)
			}
		}
	}()

	rand.Seed(time.Now().UnixNano())
	random := utils.RandomIntRange(1, 5)
	ticker := time.NewTicker(time.Duration(random) * time.Second)
	defer ticker.Stop()
	count := 0
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			send(c, count)
			count++
			if count >= 3 {
				close(stopReading)
				time.Sleep(100 * time.Millisecond)
				err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				if err != nil {
					panic(err)
				}
				wg.Done()
				return
			}
		}
	}
}

func connect() *websocket.Conn {
	u := url.URL{Scheme: "ws", Host: "localhost:9001", Path: "/ws"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial error:", err)
	}
	return c
}

func send(conn *websocket.Conn, count int) {
	data := &pbgo.CSSendMsgReq{
		Msg: conn.LocalAddr().String() + "_" + fmt.Sprintf("%d", count),
	}
	msgData, msgInfo, _ := message.EncodeMessage(data)
	// 使用加密的方式发送信息
	//encryptStr, _ := encryption.RSAEncrypt(msgData, encryption.RSAWSPublicKey)
	//mb := &socket.MsgBase{
	//	MsgId:  msgInfo.ID,
	//	FlagId: 1,
	//}
	//buf := mb.MarshalBytes(encryptStr)

	// 不使用加密的方式发送信息
	mb := &message.MsgBase{
		MsgId: msgInfo.ID,
	}
	buf := mb.MarshalBytes(msgData)
	conn.WriteMessage(websocket.BinaryMessage, buf)
}
