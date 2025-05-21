package tcp

import (
	"fmt"
	"github.com/ljhe/scream/core"
	"github.com/ljhe/scream/core/iface"
	"github.com/ljhe/scream/core/socket"
	"github.com/ljhe/scream/core/socket/sessions"
	"log"
	"net"
	"sync"
	"time"
)

type tcpConnector struct {
	socket.RuntimeTag                              // 节点运行状态相关
	socket.TCPSocketOption                         // socket相关设置
	socket.Processor                               // 事件处理相关
	socket.ServerNodeProperty                      // 节点配置属性相关
	socket.ContextSet                              // 节点上下文相关
	iface.ISessionManager                          // 会话管理
	session                   *sessions.TCPSession // 连接会话
	wg                        sync.WaitGroup
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
	return core.SocketTypTcpConnector
}

func init() {
	socket.RegisterServerNode(func() iface.INetNode {
		node := new(tcpConnector)
		node.ISessionManager = sessions.NewSessionManager()
		node.session = sessions.NewTcpSession(nil, node)
		node.TCPSocketOption.Init()
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
		log.Printf("connect success. addr:%v time:%d \n", t.GetAddr(), time.Now().Unix())
		t.wg.Add(1)
		t.session.SetConn(conn)
		t.session.Start()
		// 连接事件
		t.ProcEvent(&socket.RcvMsgEvent{Sess: t.session, Message: &socket.SessionConnected{}})
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
