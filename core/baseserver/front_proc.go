package baseserver

import (
	"errors"
	"fmt"
	"github.com/ljhe/scream/core/iface"
)

var ErrUserHasBeenBind = errors.New(fmt.Sprintf("user has been bind"))

func CreateUser(cliSession iface.ISession, openId, platform string) *ClientUser {
	user := ClientUserManager.AddClient(cliSession, openId, platform)
	// 绑定到对应的session上 一个session对应一个玩家
	cliSession.(iface.IContextSet).SetContextData("user", user)
	return user
}

func SessionUser(cliSession iface.ISession) *ClientUser {
	if cliSession == nil {
		return nil
	}
	if data, ok := cliSession.(iface.IContextSet).GetContextData("user"); ok {
		if data == nil {
			return nil
		}
		return data.(*ClientUser)
	}
	return nil
}

//func BindUser(cliSession iface.ISession, openId, platform string) error {
//
//	return nil
//}

// BindClient 绑定客户端连接到服务器
func BindClient(cliSession iface.ISession, openId, platform string) (*ClientUser, error) {
	cliUser := SessionUser(cliSession)
	// 用户已经绑定
	if cliUser != nil {
		return nil, ErrUserHasBeenBind
	}

	cliUser = CreateUser(cliSession, openId, platform)
	return cliUser, nil
}
