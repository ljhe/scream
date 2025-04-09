package baseserver

import "common/iface"

var ClientUserManager = NewClientUserManager()

type ClientUserManagerModel struct {
}

func NewClientUserManager() *ClientUserManagerModel {
	return &ClientUserManagerModel{}
}

func (c *ClientUserManagerModel) AddClient(cliSession iface.ISession, openId, platform string) *ClientUser {
	cli := NewUser(cliSession)
	cli.OpenId = openId
	cli.Platform = platform
	return cli
}
