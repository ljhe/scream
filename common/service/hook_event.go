package service

import (
	"common"
	"common/baseserver"
	"common/iface"
	"common/socket"
	"common/util"
	"pbgo"
	plugins "plugins/etcd"
	"plugins/logrus"
	"reflect"
	"time"
)

type ServerEventHook struct {
}

func (eh *ServerEventHook) InEvent(iv iface.IProcEvent) iface.IProcEvent {
	switch msg := iv.Msg().(type) {
	case *socket.SessionAccepted:
		logrus.Log(logrus.LogsSystem).Printf("receive SessionAccepted success. session:%d \n", iv.Session().GetId())
		return nil
	case *socket.SessionConnected:
		// 从内存中的etcd获取服务器信息
		ctx := iv.Session().Node().(common.ContextSet)
		var ed *plugins.ETCDServiceDesc
		if ctx.RawContextData(common.ContextSetEtcdKey, &ed) {
			prop := iv.Session().Node().(common.ServerNodeProperty)
			// 连接上服务器节点后 发送确认信息 告诉对端自己的服务器信息
			iv.Session().Send(&pbgo.ServiceIdentifyACK{
				ServiceId:       util.GenServiceId(prop),
				ServiceName:     prop.GetName(),
				ServerStartTime: util.GetCurrentTimeMs(),
			})
			// 添加远程的服务器节点信息到本地
			baseserver.AddServiceNode(iv.Session(), ed.Id, ed.Name, "local")
			logrus.Log(logrus.LogsSystem).Printf("send ServiceIdentifyACK [%v]->[%v] sessionId=%v \n",
				util.GenServiceId(prop), ed.Id, iv.Session().GetId())
		} else {
			logrus.Log(logrus.LogsSystem).Println("connector connect err. etcd not exist", msg)
		}
		return nil
	case *pbgo.ServiceIdentifyACK:
		// 来自其他服务器的连接确认信息
		logrus.Log(logrus.LogsSystem).Printf("receive ServiceIdentifyACK from [%v]  sessionId:%v \n", msg.ServiceId, iv.Session().GetId())
		// 重连时会有问题 重连上来时 但是上一个连接还未移除(正在移除中) 导致重连失败(想连接的没连接上 该移除的正在移除)
		// 通过PingReq超时断开连接 来触发断线重连
		if serviceNode := baseserver.GetServiceNode(msg.ServiceId); serviceNode == nil {
			// 添加连接上来的对端服务
			baseserver.AddServiceNode(iv.Session(), msg.ServiceId, msg.ServiceName, "remote")
			// 服务器之间的心跳检测
			// acceptor触发send connector触发rcv
			// 所以这里只能反应acceptor端的send和connector端的rcv是否正常
			iv.Session().HeartBeat(&pbgo.PingReq{Ms: time.Now().UnixMilli()})
		}
		return nil
	case *pbgo.PingReq:
		// 来自ServiceIdentifyACK接收端的服务器信息
		ctx := iv.Session().(common.ContextSet)
		var ed *plugins.ETCDServiceDesc
		iv.Session().IncRcvPingNum(1)
		if iv.Session().RcvPingNum() >= 10 {
			iv.Session().IncRcvPingNum(-1)
			if ctx.RawContextData(common.ContextSetCtxKey, &ed) {
				logrus.Log(logrus.LogsSystem).Printf("receive PingReq from [%v] session=%v \n", ed.Id, iv.Session().GetId())
			}
		}
		iv.Session().Send(&pbgo.PingAck{Ms: time.Now().UnixMilli()})
		return nil
	case *pbgo.PingAck:
		ctx := iv.Session().(common.ContextSet)
		var ed *plugins.ETCDServiceDesc
		iv.Session().IncRcvPingNum(1)
		if iv.Session().RcvPingNum() >= 10 {
			iv.Session().IncRcvPingNum(-1)
			if ctx.RawContextData(common.ContextSetCtxKey, &ed) {
				logrus.Log(logrus.LogsSystem).Printf("receive PingAck from [%v] session=%v \n", ed.Id, iv.Session().GetId())
			}
		}
		return nil
	case *socket.SessionClosed:
		sid := baseserver.RemoveServiceNode(iv.Session())
		logrus.Log(logrus.LogsSystem).Printf("SessionClosed sessionId=%v sid=%v \n", iv.Session().GetId(), sid)
		return nil
	default:
		logrus.Log(logrus.LogsSystem).Printf("receive unknown msg %v msgT:%v ivM %v \n", msg, reflect.TypeOf(msg), iv.Msg())
	}
	return iv
}

func (eh *ServerEventHook) OutEvent(ov iface.IProcEvent) iface.IProcEvent {
	return ov
}

type WsEventHook struct {
}

func (wh *WsEventHook) InEvent(iv iface.IProcEvent) iface.IProcEvent {
	switch msg := iv.Msg().(type) {
	case *pbgo.CSPingReq:
		iv.Session().Send(&pbgo.SCPingAck{})
		return nil
	case *socket.SessionClosed:
		e := iv.(*common.RcvMsgEvent)
		if e.Err != nil {
			logrus.Log(logrus.LogsSystem).Infof("ws session closed. err:%v", e.Err)
		}
		return nil
	case *socket.SessionAccepted:
		logrus.Log(logrus.LogsSystem).Infof("WS-SessionConnected cliId=%v", iv.Session().GetId())
		return nil
	case *pbgo.CSSendMsgReq:
		m := iv.Msg().(*pbgo.CSSendMsgReq)
		iv.Session().Send(&pbgo.SCSendMsgAck{Msg: m.Msg})
		return nil
	case *pbgo.CSLoginReq:
		m := iv.Msg().(*pbgo.CSLoginReq)
		cliUser, err := baseserver.BindClient(iv.Session(), m.OpenId, m.Platform)
		if err == nil {
			// 绑定成功 转发给对应的服务器做处理
			node, _ := baseserver.GetServiceNodeAndSession("", common.ServiceNodeTypeGameStr, 0)
			err = cliUser.ClientDirect2Backend(node, 0, 0, []byte(m.OpenId), common.ServiceNodeTypeGameStr)
			if err != nil {
				return nil
			}
			iv.Session().Send(&pbgo.SCLoginAck{Error: int32(pbgo.ErrorCode_ERROR_OK)})
		} else {
			logrus.Log(logrus.LogsSystem).Errorf("CSLoginReq BindClient err:%s. openId=%s", err, m.OpenId)
			iv.Session().Send(&pbgo.SCLoginAck{Error: int32(pbgo.ErrorCode_ERROR_SESSION_BIND_CLIENT)})
		}
		return nil
	default:
		logrus.Log(logrus.LogsSystem).Infof("receive unknown msg %v msgT:%v ivM %v \n", msg, reflect.TypeOf(msg), iv.Msg())
	}
	return iv
}

func (wh *WsEventHook) OutEvent(ov iface.IProcEvent) iface.IProcEvent {
	return ov
}
