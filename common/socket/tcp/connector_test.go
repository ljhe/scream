package tcp

import (
	"common"
	"common/service"
	"common/socket"
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestNewConnector(t *testing.T) {
	node := socket.NewServerNode(common.SocketTypTcpConnector, "test", "0.0.0.0:2701")
	rand.New(rand.NewSource(time.Now().UnixNano()))
	n := rand.Intn(10)
	for i := 0; i <= n; i++ {
		node.Start()
	}
	fmt.Println("node num:", n+1)
	service.WaitExitSignal()
}
