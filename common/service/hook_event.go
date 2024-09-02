package service

import (
	"common"
	"common/iface"
	plugins "common/plugins/etcd"
	"common/util"
	"fmt"
	"log"
)

type ServerEventHook struct {
}

func (eh *ServerEventHook) InEvent(iv iface.IProcEvent) iface.IProcEvent {
	switch msg := iv.Msg().(type) {
	case *common.SessionAccepted:
		// 服务器之间的心跳检测
		// acceptor触发send connector触发rcv
		// 所以这里只能反应acceptor端的send和connector端的rcv是否正常
		iv.Session().HeartBeat(fmt.Sprintf("server ping req"))
		return nil
	case *common.SessionConnected:
		// TODO 连接上服务器节点后 发送确认信息 告诉对端自己的服务器信息
		ctx := iv.Session().Node().(common.ContextSet)
		var ed *plugins.ETCDServiceDesc
		// 从内存中的etcd获取服务器信息
		if ctx.RawContextData(ContextSetEtcdKey, &ed) {
			// TODO 把服务器节点添加到本地
			log.Printf("send ServiceIdentifyACK [%v]->[%v] sessionId=%v",
				util.GenServiceId(iv.Session().Node().(common.ServerNodeProperty)), ed.Id, iv.Session().GetId())
		} else {
			log.Println("connector connect err. etcd not exist", msg)
		}
		return nil
	}
	return iv
}

func (eh *ServerEventHook) OutEvent(ov iface.IProcEvent) iface.IProcEvent {
	return ov
}
