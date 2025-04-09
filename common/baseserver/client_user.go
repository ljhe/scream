package baseserver

import (
	"common/iface"
	"common/plugins/logrus"
)

type ClientUser struct {
	ClientSession iface.ISession

	// 登录验证使用
	OpenId   string
	Platform string
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
	logrus.Log(logrus.LogsSystem).Infof("bind success, send other server...")
}
