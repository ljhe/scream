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
	socket.NetRuntimeTag                     // 节点运行状态相关
	socket.NetTCPSocketOption                // socket相关设置
	socket.NetProcessorRPC                   // 事件处理相关
	socket.NetServerNodeProperty             // 节点配置属性相关
	socket.NetContextSet                     // 节点上下文相关
	socket.SessionManager                    // 会话管理
	session                      *tcpSession // 连接会话
	wg                           sync.WaitGroup
}

func (t *tcpConnector) Start() iface.INetNode {
	// 正在停止的话 需要先等待
	t.StopWg.Wait()
	if t.GetRunState() {
		return t
	}
	go t.connect()
	return t
}

func (t *tcpConnector) Stop() {
	if !t.GetRunState() {
		return
	}
	t.SetCloseFlag(true)
	t.StopWg.Add(1)
	t.session.Close()
	t.wg.Done()
	t.StopWg.Wait()
	log.Println("tcp connector stop success.")
}

func (t *tcpConnector) GetTyp() string {
	return common.SocketTypTcpConnector
}

func init() {
	socket.RegisterServerNode(func() iface.INetNode {
		node := new(tcpConnector)
		node.SessionManager = socket.NewNetSessionManager()
		node.session = newTcpSession(nil, node)
		node.NetTCPSocketOption.Init()
		return node
	})
	log.Println("tcp connector register success.")
}

func (t *tcpConnector) connect() {
	t.SetRunState(true)
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
		t.session.SetConn(conn)
		t.session.Start()
		// 连接事件
		t.ProcEvent(&common.RcvMsgEvent{Sess: t.session, Message: &socket.SessionConnected{}})
		//go t.deal(conn)
		t.wg.Wait()
		if t.GetCloseFlag() {
			break
		}
	}
	t.SetRunState(false)
	t.StopWg.Done()
	log.Println("tcp connector close.")
}

func (t *tcpConnector) deal(conn net.Conn) {
	buffer := make([]byte, 1024)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			break
		}
		fmt.Println("receive: ", string(buffer[:n]))
	}
}
