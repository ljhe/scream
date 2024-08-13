package tcp

import (
	"common"
	"common/iface"
	"common/socket"
	"context"
	"log"
	"net"
	"syscall"
	"time"
)

type tcpAcceptor struct {
	socket.NetRuntimeTag         // 节点运行状态相关
	socket.NetProcessorRPC       // 事件处理相关
	socket.NetServerNodeProperty // 节点配置属性相关
	listener                     net.Listener
}

func (t *tcpAcceptor) Start() iface.INetNode {
	listenConfig := net.ListenConfig{
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
	ln, err := listenConfig.Listen(context.Background(), "tcp", t.GetAddr())
	if err != nil {
		log.Println("tcp listen error:", err)
		return nil
	}
	t.listener = ln
	log.Printf("tcp listen success. addr:%v \n", t.GetAddr())
	go t.tcpAccept()
	return t
}

func (t *tcpAcceptor) Stop() {
	t.SetCloseFlag(true)
	t.listener.Close()
	log.Println("tcp acceptor stop success.")
}

func (t *tcpAcceptor) GetTyp() string {
	return common.SocketTypTcpAcceptor
}

func init() {
	socket.RegisterServerNode(func() iface.INetNode {
		return &tcpAcceptor{}
	})
	log.Println("tcp acceptor register success.")
}

func (t *tcpAcceptor) tcpAccept() {
	for {
		//conn, err := t.listener.Accept()
		_, err := t.listener.Accept()
		// 判断节点是否关闭
		if t.GetCloseFlag() {
			break
		}
		if err != nil {
			// 尝试重连
			if opErr, ok := err.(net.Error); ok && opErr.Temporary() {
				select {
				case <-time.After(time.Second * 3):
					continue
				}
			}
			log.Println("tcp accept error:", err)
			break
		}
		/////go t.deal(conn)
		t.ProcEvent(&common.ReceiveMsgEvent{Message: "这里是测试!!"})
	}
	log.Println("tcp acceptor break.")
}

//func (t *tcpAcceptor) deal(conn net.Conn) {
//	buffer := make([]byte, 1024)
//	for {
//		n, err := conn.Read(buffer)
//		if err != nil {
//			break
//		}
//		fmt.Println("收到消息: ", string(buffer[:n]))
//	}
//}
