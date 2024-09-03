package service

import (
	"common"
	"common/iface"
	plugins "common/plugins/etcd"
	"common/util"
	"log"
	"reflect"
)

type ServerEventHook struct {
}

func (eh *ServerEventHook) InEvent(iv iface.IProcEvent) iface.IProcEvent {
	switch msg := iv.Msg().(type) {
	case *common.SessionAccepted:
		// 服务器之间的心跳检测
		// acceptor触发send connector触发rcv
		// 所以这里只能反应acceptor端的send和connector端的rcv是否正常
		//iv.Session().HeartBeat(fmt.Sprintf("server ping req"))
		return nil
	case *common.SessionConnected:
		// 从内存中的etcd获取服务器信息
		ctx := iv.Session().Node().(common.ContextSet)
		var ed *plugins.ETCDServiceDesc
		if ctx.RawContextData(ContextSetEtcdKey, &ed) {
			prop := iv.Session().Node().(common.ServerNodeProperty)
			// 连接上服务器节点后 发送确认信息 告诉对端自己的服务器信息
			iv.Session().Send(&common.ServiceIdentifyACK{
				ServiceId:       util.GenServiceId(prop),
				ServiceName:     prop.GetName(),
				ServerStartTime: util.GetCurrentTimeMs(),
			})
			// TODO 把服务器节点添加到本地
			log.Printf("send ServiceIdentifyACK [%v]->[%v] sessionId=%v \n",
				util.GenServiceId(prop), ed.Id, iv.Session().GetId())
		} else {
			log.Println("connector connect err. etcd not exist", msg)
		}
		return nil
	case *common.ServiceIdentifyACK:
		// 来自其他服务器的连接确认信息
		log.Printf("receive ServiceIdentifyACK from [%v]  sessionId:%v \n", msg.ServiceId, iv.Session().GetId())
		return nil
	default:
		log.Printf("receive unknown msg %v msgT:%v ivM %v \n", msg, reflect.TypeOf(msg), iv.Msg())
	}
	return iv
}

func (eh *ServerEventHook) OutEvent(ov iface.IProcEvent) iface.IProcEvent {
	return ov
}
