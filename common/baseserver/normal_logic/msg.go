package normal_logic

import (
	"common/baseserver"
	"common/iface"
)

func HandleMessage(userHandler func(ev iface.IProcEvent, cliID baseserver.ClientID)) func(ev iface.IProcEvent) {
	return func(e iface.IProcEvent) {
		userHandler(nil, baseserver.ClientID{})
	}
}
