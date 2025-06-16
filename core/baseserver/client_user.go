package baseserver

import (
	"errors"
	"fmt"
	"github.com/ljhe/scream/3rd/logrus"
	"github.com/ljhe/scream/core/iface"
	"github.com/ljhe/scream/pbgo"
)

type ClientUser struct {
	ClientSession iface.ISession

	// 登录验证使用
	OpenId   string
	Platform string
}

type ClientID struct {
	SessID     uint64 // 客户端在网管上的sessionId
	ServiceID  string // 客户端所在的网关
	SessIdList []uint64
}

func NewUser(cliSession iface.ISession) *ClientUser {
	cli := &ClientUser{
		ClientSession: cliSession,
	}
	cli.init()
	return cli
}

func (cli *ClientUser) init() {
	// TODO 注册状态机
}

func (cli *ClientUser) AddClient(cliSession iface.ISession, openId, platform string) *ClientUser {
	u := NewUser(cliSession)
	u.OpenId = openId
	u.Platform = platform
	return u
}

// SendServer gate把接收到的数据直接发送到后端其他服务器节点
func (cli *ClientUser) SendServer() {
	logrus.Infof("bind success, send other server...")
}

// gate把接收到的数据直接发送到后端服务器节点
func (this *ClientUser) ClientDirect2Backend(serviceId string, msgId int, seqId uint32, msgData []byte, serviceType string) error {
	// 获得后端服务器节点，并发送
	service := GetServiceNode(serviceId)
	if service == nil {
		return errors.New(fmt.Sprintf("server nod not find ClientDirect2Backend:%v %v %v", serviceType, msgId, serviceId))
	}

	// 用户ID绑定处理
	service.Send(&pbgo.CSLoginReq{OpenId: string(msgData)})
	return nil
}
