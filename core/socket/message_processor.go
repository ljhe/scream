package socket

import "github.com/ljhe/scream/core/iface"

type TCPMessageProcessor struct {
}

func (tp *TCPMessageProcessor) OnRcvMsg(s iface.ISession) (msg interface{}, err error) {
	opt := s.Node().(Option)
	opt.SocketReadTimeout(s, func() {
		p := TcpDataPacket{}
		msg, err = p.ReadMessage(s)
	})
	return
}

func (tp *TCPMessageProcessor) OnSendMsg(s iface.ISession, msg interface{}) (err error) {
	opt := s.Node().(Option)
	opt.SocketWriteTimeout(s, func() {
		p := TcpDataPacket{}
		err = p.SendMessage(s, msg)
	})
	return err
}

type WSMessageProcessor struct {
}

func (tp *WSMessageProcessor) OnRcvMsg(s iface.ISession) (msg interface{}, err error) {
	opt := s.Node().(Option)
	opt.SocketReadTimeout(s, func() {
		p := WsDataPacket{}
		msg, err = p.ReadMessage(s)
	})
	return
}

func (tp *WSMessageProcessor) OnSendMsg(s iface.ISession, msg interface{}) (err error) {
	opt := s.Node().(Option)
	opt.SocketWriteTimeout(s, func() {
		p := WsDataPacket{}
		err = p.SendMessage(s, msg)
	})
	return err
}
