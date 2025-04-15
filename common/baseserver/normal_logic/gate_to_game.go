package normal_logic

import (
	"common/baseserver"
	"common/iface"
	"log"
	"pbgo"
)

func init() {
	pbgo.Handle_GAME_CSLoginReq = HandleMessage(func(ev iface.IProcEvent, cliId baseserver.ClientID) {
		log.Println("CSLoginReq implements")
	})
}
