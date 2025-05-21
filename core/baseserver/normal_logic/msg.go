package normal_logic

import (
	"github.com/ljhe/scream/core/baseserver"
	"github.com/ljhe/scream/core/iface"
)

func HandleMessage(userHandler func(ev iface.IProcEvent, cliID baseserver.ClientID)) func(ev iface.IProcEvent) {
	return func(e iface.IProcEvent) {
		userHandler(nil, baseserver.ClientID{})
	}
}
