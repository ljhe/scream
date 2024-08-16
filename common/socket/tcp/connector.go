package tcp

import (
	"common"
	"common/iface"
	"common/socket"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

type tcpConnector struct {
	socket.NetRuntimeTag         // 节点运行状态相关
	socket.NetServerNodeProperty // 节点配置属性相关
	wg                           sync.WaitGroup
}

func (t *tcpConnector) Start() iface.INetNode {
	go t.connect()
	return t
}

func (t *tcpConnector) Stop() {
	t.wg.Done()
	log.Println("tcp connector stop success.")
}

func (t *tcpConnector) GetTyp() string {
	return common.SocketTypTcpConnector
}

func init() {
	socket.RegisterServerNode(func() iface.INetNode {
		return &tcpConnector{}
	})
	log.Println("tcp connector register success.")
}

func (t *tcpConnector) connect() {
	t.SetCloseFlag(true)
	for {
		conn, err := net.Dial("tcp", t.GetAddr())
		if err != nil {
			if t.GetCloseFlag() {
				fmt.Printf("connect error:%v \n", err)
				break
			}
			// 连接失败后 重连
			select {
			case <-time.After(time.Second * 3):
				continue
			}
		}
		fmt.Printf("connect success. addr:%v time:%d \n", t.GetAddr(), time.Now().Unix())
		t.wg.Add(1)
		_, err = conn.Write([]byte("这里是一条测试数据"))
		if err != nil {
			fmt.Println("send data error")
			conn.Close()
		}
		t.wg.Wait()
	}
}
