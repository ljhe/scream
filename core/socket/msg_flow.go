package socket

import "github.com/ljhe/scream/core/iface"

type TCPMsgFlow struct {
}

func (tp *TCPMsgFlow) OnRcvMsg(s iface.ISession) (msg interface{}, err error) {
	opt := s.Node().(Option)
	opt.SocketReadTimeout(s, func() {
		p := TcpDataPacket{}
		msg, err = p.ReadMessage(s)
	})
	return
}

func (tp *TCPMsgFlow) OnSendMsg(s iface.ISession, msg interface{}) (err error) {
	opt := s.Node().(Option)
	opt.SocketWriteTimeout(s, func() {
		p := TcpDataPacket{}
		err = p.SendMessage(s, msg)
	})
	return err
}

type WSMsgFlow struct {
}

func (tp *WSMsgFlow) OnRcvMsg(s iface.ISession) (msg interface{}, err error) {
	opt := s.Node().(Option)
	opt.SocketReadTimeout(s, func() {
		p := WsDataPacket{}
		msg, err = p.ReadMessage(s)
	})
	return
}

func (tp *WSMsgFlow) OnSendMsg(s iface.ISession, msg interface{}) (err error) {
	opt := s.Node().(Option)
	opt.SocketWriteTimeout(s, func() {
		p := WsDataPacket{}
		err = p.SendMessage(s, msg)
	})
	return err
}
