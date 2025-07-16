package socket

import (
	"github.com/ljhe/scream/3rd/logrus"
	"github.com/ljhe/scream/core/baseserver"
	"github.com/ljhe/scream/core/iface"
	"github.com/ljhe/scream/core/message"
	"github.com/ljhe/scream/def"
	"github.com/ljhe/scream/pbgo"
	"github.com/ljhe/scream/utils"
	"reflect"
	"time"
)

type ServerHookEvent struct {
}

func (eh *ServerHookEvent) InEvent(iv iface.IProcEvent) iface.IProcEvent {
	switch msg := iv.Msg().(type) {
	case *SessionAccepted:
		logrus.Printf("receive SessionAccepted success. session:%d", iv.Session().GetId())
		return nil
	case *SessionConnected:
		// 从内存中的etcd获取服务器信息
		ctx := iv.Session().Node().(iface.IContextSet)
		var ed *utils.ServerInfo
		if ctx.RawContextData(def.ContextSetEtcdKey, &ed) {
			prop := iv.Session().Node().(iface.INodeProp)
			// 连接上服务器节点后 发送确认信息 告诉对端自己的服务器信息
			iv.Session().Send(&pbgo.ServiceIdentifyACK{
				ServiceId:       utils.GenSelfServiceId(prop.GetName(), prop.GetServerTyp(), prop.GetIndex()),
				ServiceName:     prop.GetName(),
				ServerStartTime: utils.GetCurrentTimeMs(),
			})
			// 添加远程的服务器节点信息到本地
			baseserver.AddServiceNode(iv.Session(), ed.Id, ed.Name, "local")
			logrus.Printf("send ServiceIdentifyACK [%v]->[%v] sessionId=%v",
				utils.GenSelfServiceId(prop.GetName(), prop.GetServerTyp(), prop.GetIndex()), ed.Id, iv.Session().GetId())
		} else {
			logrus.Infof("connector connect err. etcd not exist message:%v", msg)
		}
		return nil
	case *SessionClosed:
		sid := baseserver.RemoveServiceNode(iv.Session())
		logrus.Printf("SessionClosed sessionId=%v sid=%v", iv.Session().GetId(), sid)
		return nil
	case *pbgo.ServiceIdentifyACK:
		// 来自其他服务器的连接确认信息
		logrus.Printf("receive ServiceIdentifyACK from [%v]  sessionId:%v", msg.ServiceId, iv.Session().GetId())
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
		ctx := iv.Session().(iface.IContextSet)
		var ed *utils.ServerInfo
		iv.Session().IncRcvPingNum(1)
		if iv.Session().RcvPingNum() >= 10 {
			iv.Session().IncRcvPingNum(-1)
			if ctx.RawContextData(def.ContextSetCtxKey, &ed) {
				logrus.Printf("receive PingReq from [%v] session=%v", ed.Id, iv.Session().GetId())
			}
		}
		iv.Session().Send(&pbgo.PingAck{Ms: time.Now().UnixMilli()})
		return nil
	case *pbgo.PingAck:
		ctx := iv.Session().(iface.IContextSet)
		var ed *utils.ServerInfo
		iv.Session().IncRcvPingNum(1)
		if iv.Session().RcvPingNum() >= 10 {
			iv.Session().IncRcvPingNum(-1)
			if ctx.RawContextData(def.ContextSetCtxKey, &ed) {
				logrus.Printf("receive PingAck from [%v] session=%v", ed.Id, iv.Session().GetId())
			}
		}
		return nil
	case *pbgo.MsgTransmitNtf:
		data, err := message.DecodeMessage(uint16(msg.MsgId), msg.Data)
		if err != nil {
			panic(err)
		}

		iv.Session().TransmitChild(msg.SessionId, data)
		logrus.Printf("receive MsgTransmitNtf message. main_session:%d client_session:%d dataT:%v data:%v",
			iv.Session().GetId(), msg.SessionId, reflect.TypeOf(data), data)
		return nil
	default:
		logrus.Printf("receive unknown message %v msgT:%v ivM %v sessionId:%d",
			msg, reflect.TypeOf(msg), iv.Msg(), iv.Session().GetId())
	}
	return iv
}

func (eh *ServerHookEvent) OutEvent(ov iface.IProcEvent) iface.IProcEvent {
	return ov
}

type WsHookEvent struct {
}

func (wh *WsHookEvent) InEvent(iv iface.IProcEvent) iface.IProcEvent {
	switch msg := iv.Msg().(type) {
	case *SessionAccepted:
		logrus.Infof("WS-SessionConnected cliId=%v", iv.Session().GetId())
		return nil
	case *pbgo.WSSessionClosedNtf:
		logrus.Infof("ws session closed. sessionId:%d", iv.Session().GetId())

		// 测试消息转发关闭
		node, _ := baseserver.GetServiceNodeAndSession("", def.ServiceNodeTypeGameStr, 0)
		service := baseserver.GetServiceNode(node)
		if service == nil {
			return nil
		}
		// 服务器间通信 增加特有结构体 里面包含sessionId
		bytes, info, err := message.EncodeMessage(&pbgo.WSSessionClosedNtf{})
		if err != nil {
			panic(err)
		}
		service.Send(&pbgo.MsgTransmitNtf{
			SessionId: iv.Session().GetId(),
			MsgId:     uint32(info.ID),
			Data:      bytes,
		})

		// 关闭客户端到ws的发送端
		iv.Session().Close()
		return nil
	case *pbgo.CSPingReq:
		iv.Session().Send(&pbgo.SCPingAck{})
		return nil
	case *pbgo.CSSendMsgReq:
		m := iv.Msg().(*pbgo.CSSendMsgReq)
		logrus.Infof("receive client message. sessionId:%d message:%v", iv.Session().GetId(), m.Msg)

		// 测试消息转发
		node, _ := baseserver.GetServiceNodeAndSession("", def.ServiceNodeTypeGameStr, 0)
		service := baseserver.GetServiceNode(node)
		if service == nil {
			return nil
		}
		// 服务器间通信 增加特有结构体 里面包含sessionId
		bytes, info, err := message.EncodeMessage(iv.Msg())
		if err != nil {
			panic(err)
		}
		service.Send(&pbgo.MsgTransmitNtf{
			SessionId: iv.Session().GetId(),
			MsgId:     uint32(info.ID),
			Data:      bytes,
		})

		// 返回给客户端消息
		iv.Session().Send(&pbgo.SCSendMsgAck{Msg: m.Msg})
		return nil
	case *pbgo.CSLoginReq:
		m := iv.Msg().(*pbgo.CSLoginReq)
		cliUser, err := baseserver.BindClient(iv.Session(), m.OpenId, m.Platform)
		if err == nil {
			// 绑定成功 转发给对应的服务器做处理
			node, _ := baseserver.GetServiceNodeAndSession("", def.ServiceNodeTypeGameStr, 0)
			err = cliUser.ClientDirect2Backend(node, 0, 0, []byte(m.OpenId), def.ServiceNodeTypeGameStr)
			if err != nil {
				return nil
			}
			iv.Session().Send(&pbgo.SCLoginAck{Error: int32(pbgo.ErrorCode_ERROR_OK)})
		} else {
			logrus.Errorf("CSLoginReq BindClient err:%s. openId=%s", err, m.OpenId)
			iv.Session().Send(&pbgo.SCLoginAck{Error: int32(pbgo.ErrorCode_ERROR_SESSION_BIND_CLIENT)})
		}
		return nil
	default:
		logrus.Infof("receive unknown message %v msgT:%v ivM %v", msg, reflect.TypeOf(msg), iv.Msg())
	}
	return iv
}

func (wh *WsHookEvent) OutEvent(ov iface.IProcEvent) iface.IProcEvent {
	return ov
}

type SessionChildHookEvent struct{}

func (sc *SessionChildHookEvent) InEvent(iv iface.IProcEvent) iface.IProcEvent {
	s, ok := iv.Session().(iface.ISessionChild)
	if !ok {
		panic("")
	}
	switch msg := iv.Msg().(type) {
	case *pbgo.WSSessionClosedNtf:
		logrus.Infof("SessionChildHookEvent session closed. sessionId:%d", s.GetSessionId())
		iv.Session().DelChild(s.GetSessionId())
		return nil
	default:
		logrus.Infof("receive unknown message %v msgT:%v ivM %v", msg, reflect.TypeOf(msg), iv.Msg())
	}
	return iv
}

func (sc *SessionChildHookEvent) OutEvent(ov iface.IProcEvent) iface.IProcEvent {
	return ov
}
