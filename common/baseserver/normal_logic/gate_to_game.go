package normal_logic

import (
	"github.com/ljhe/scream/common/baseserver"
	"github.com/ljhe/scream/common/iface"
	"github.com/ljhe/scream/pbgo"
	"log"
)

func init() {
	pbgo.Handle_GAME_CSLoginReq = HandleMessage(func(ev iface.IProcEvent, cliId baseserver.ClientID) {
		log.Println("CSLoginReq implements")
	})
}
