package normal_logic

import (
	"github.com/ljhe/scream/core/baseserver"
	"github.com/ljhe/scream/core/iface"
	"github.com/ljhe/scream/pbgo"
	"log"
)

func init() {
	pbgo.Handle_GAME_CSLoginReq = HandleMessage(func(ev iface.IProcEvent, cliId baseserver.ClientID) {
		log.Println("CSLoginReq implements")
	})
	pbgo.Handle_GAME_CSSendMsgReq = HandleMessage(func(ev iface.IProcEvent, cliId baseserver.ClientID) {
		log.Println("CSSendMsgReq implements")
	})
}
