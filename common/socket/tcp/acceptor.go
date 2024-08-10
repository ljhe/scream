package tcp

import (
	"common"
	"common/iface"
	"common/socket"
	"context"
	"fmt"
	"log"
	"net"
	"syscall"
)

type tcpAcceptor struct {
	socket.NetServerNodeProperty
	listener net.Listener
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
		conn, err := t.listener.Accept()
		if err != nil {
			log.Println("tcp accept error:", err)
			break
		}
		go t.deal(conn)
	}
}

func (t *tcpAcceptor) deal(conn net.Conn) {
	buffer := make([]byte, 1024)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			break
		}
		fmt.Println("收到消息: ", string(buffer[:n]))
	}
}
