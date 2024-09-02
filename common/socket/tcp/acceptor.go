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
	"time"
)

type tcpAcceptor struct {
	socket.NetRuntimeTag         // 节点运行状态相关
	socket.NetTCPSocketOption    // socket相关设置
	socket.NetProcessorRPC       // 事件处理相关
	socket.NetServerNodeProperty // 节点配置属性相关
	socket.NetContextSet         // 节点上下文相关
	socket.SessionManager        // 会话管理
	listener                     net.Listener
}

func (t *tcpAcceptor) Start() iface.INetNode {
	// 正在停止的话 需要先等待
	t.StopWg.Wait()
	// 防止重入导致错误
	if t.GetRunState() {
		return t
	}

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
		log.Println(fmt.Sprintf("tcp listen error:%v. addr:%v", err, t.GetAddr()))
		return nil
	}
	t.listener = ln
	log.Printf("tcp listen success. addr:%v \n", t.GetAddr())
	go t.tcpAccept()
	return t
}

func (t *tcpAcceptor) Stop() {
	if !t.GetRunState() {
		return
	}
	// 添加结束协程
	t.StopWg.Add(1)
	// 设置结束标签
	t.SetCloseFlag(true)
	t.listener.Close()
	// 等待协程结束
	t.StopWg.Wait()
	log.Println("tcp acceptor stop success.")
}

func (t *tcpAcceptor) GetTyp() string {
	return common.SocketTypTcpAcceptor
}

func init() {
	socket.RegisterServerNode(func() iface.INetNode {
		return &tcpAcceptor{
			SessionManager: socket.NewNetSessionManager(),
		}
	})
	log.Println("tcp acceptor register success.")
}

func (t *tcpAcceptor) tcpAccept() {
	t.SetRunState(true)
	for {
		conn, err := t.listener.Accept()
		// 判断节点是否关闭
		if t.GetCloseFlag() {
			break
		}
		if err != nil {
			// 尝试重连
			if opErr, ok := err.(net.Error); ok && opErr.Temporary() {
				select {
				case <-time.After(time.Millisecond * 3):
					continue
				}
			}
			log.Println("tcp accept error:", err)
			break
		}
		log.Println("tcp accept success. remoteAddr:", conn.RemoteAddr())
		//go t.deal(conn)
		func() {
			session := newTcpSession(conn, t)
			session.Start()
			// 通知上层主事件 (将回调放入队列中 防止多线程冲突)
			t.ProcEvent(&common.RcvMsgEvent{Sess: session, Message: &common.SessionAccepted{}})
		}()
	}
	t.SetRunState(false)
	t.SetCloseFlag(false)
	t.StopWg.Done()
	log.Println("tcp acceptor close.")
}

func (t *tcpAcceptor) deal(conn net.Conn) {
	buffer := make([]byte, 1024)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			break
		}
		fmt.Println("receive: ", string(buffer[:n]))
		conn.Write([]byte("handshakes ack."))
	}
}
