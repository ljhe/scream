package socket

import (
	"common"
	"common/iface"
	"common/plugins/logrus"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net"
)

type NetProcessorRPC struct {
	MsgRPC    common.MessageProcessor // 根据不同对象来处理消息读写的加解密
	Hooker    common.EventHook        // 不进入主消息队列 直接操作
	MsgHandle common.IMsgHandle
}

func (n *NetProcessorRPC) SetMessageProc(v common.MessageProcessor) {
	n.MsgRPC = v
}

func (n *NetProcessorRPC) SetHooker(v common.EventHook) {
	n.Hooker = v
}

func (n *NetProcessorRPC) SetMsgHandle(v common.IMsgHandle) {
	n.MsgHandle = v
}

func (n *NetProcessorRPC) GetRPC() *NetProcessorRPC {
	return n
}

func (n *NetProcessorRPC) ProcEvent(e iface.IProcEvent) {
	if n.Hooker != nil {
		e = n.Hooker.InEvent(e)
	}
	if e != nil {
		n.MsgHandle.PostCb(func() {
			fmt.Println("这里是测试数据")
		})
	}
}

func (n *NetProcessorRPC) ReadMsg(s iface.ISession) (interface{}, error) {
	if n.MsgRPC != nil {
		return n.MsgRPC.OnRcvMsg(s)
	}
	return nil, fmt.Errorf("msg rpc is nil")
}

func (n *NetProcessorRPC) SendMsg(e iface.IProcEvent) error {
	if n.Hooker != nil {
		e = n.Hooker.OutEvent(e)
	}
	if n.MsgRPC != nil {
		return n.MsgRPC.OnSendMsg(e.Session(), e.Msg())
	}
	return nil
}

type TCPMessageProcessor struct {
}

func (tp *TCPMessageProcessor) OnRcvMsg(s iface.ISession) (msg interface{}, err error) {
	reader, ok := s.Raw().(io.Reader)
	if !ok || reader == nil {
		log.Println("[TCPMessageProcessor] OnRcvMsg err")
		return nil, fmt.Errorf("[TCPMessageProcessor] OnRcvMsg err")
	}
	opt := s.Node().(Option)
	opt.SocketReadTimeout(reader.(net.Conn), func() {
		msg, err = ReadMessage(reader, 1024)
	})
	return
}

func (tp *TCPMessageProcessor) OnSendMsg(s iface.ISession, msg interface{}) (err error) {
	w, ok := s.Raw().(io.Writer)
	if !ok || w == nil {
		logrus.Log(logrus.LogsSystem).Errorf("[TCPMessageProcessor] OnSendMsg err")
		return fmt.Errorf("[TCPMessageProcessor] OnSendMsg err")
	}
	opt := s.Node().(Option)
	opt.SocketWriteTimeout(w.(net.Conn), func() {
		err = SendMessage(w, msg)
	})
	return err
}

type WSMessageProcessor struct {
}

func (tp *WSMessageProcessor) OnRcvMsg(s iface.ISession) (msg interface{}, err error) {
	conn, ok := s.Raw().(*websocket.Conn)
	if !ok || conn == nil {
		logrus.Log(logrus.LogsSystem).Errorf("[WSMessageProcessor] OnRcvMsg err")
		return nil, nil
	}
	typ, byte, err := conn.ReadMessage()
	if err != nil {
		return
	}

	// 打印客户端发送的数据
	logrus.Log(logrus.LogsSystem).Infof("ws acceptor receive msg:%v", string(byte))
	// 回复客户端
	if err = conn.WriteMessage(typ, byte); err != nil {
		return
	}
	return
}

func (tp *WSMessageProcessor) OnSendMsg(s iface.ISession, msg interface{}) error {
	return nil
}
