package pbgo

import (
	"common"
	"common/iface"
	"log"
	"reflect"
)

func registerInfo(id uint16, msgType reflect.Type) {
	RegisterMessageInfo(&MessageInfo{ID: id, Codec: GetCodec(), Type: msgType})
}

//GATE
var (
	Handle_GATE_SCLoginAck  = func(e  iface.IProcEvent){panic("SCLoginAck not implements")}
	Handle_GATE_Default		func(e  iface.IProcEvent)
)

//GAME
var (
	Handle_GAME_CSLoginReq  = func(e  iface.IProcEvent){panic("CSLoginReq not implements")}
	Handle_GAME_Default		func(e  iface.IProcEvent)
)

func GetMessageHandler(sreviceName string) common.EventCallBack {
	switch sreviceName { //note.serviceName must be lower words
	case "gate":	//GATE message process part
		return func(e iface.IProcEvent) {
			switch e.Msg().(type) {
			case *SCLoginAck: Handle_GATE_SCLoginAck(e)
			default:
				if Handle_GATE_Default != nil {
					Handle_GATE_Default(e)
				}
			}
		}

	case "game":	//GAME message process part
		return func(e iface.IProcEvent) {
			switch e.Msg().(type) {
			case *CSLoginReq: Handle_GAME_CSLoginReq(e)
			default:
				if Handle_GAME_Default != nil {
					Handle_GAME_Default(e)
				}
			}
		}

	default: 
		return nil
	}
}

func init() {
	// 协议注册
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	registerInfo(1, reflect.TypeOf((*ServiceIdentifyACK)(nil)).Elem())
	registerInfo(2, reflect.TypeOf((*PingReq)(nil)).Elem())
	registerInfo(3, reflect.TypeOf((*PingAck)(nil)).Elem())
	registerInfo(4, reflect.TypeOf((*CSPingReq)(nil)).Elem())
	registerInfo(5, reflect.TypeOf((*SCPingAck)(nil)).Elem())
	registerInfo(6, reflect.TypeOf((*CSSendMsgReq)(nil)).Elem())
	registerInfo(7, reflect.TypeOf((*SCSendMsgAck)(nil)).Elem())
	registerInfo(1000, reflect.TypeOf((*CSLoginReq)(nil)).Elem())
	registerInfo(1001, reflect.TypeOf((*SCLoginAck)(nil)).Elem())
	registerInfo(5000, reflect.TypeOf((*CSCreateRoleReq)(nil)).Elem())
	registerInfo(5001, reflect.TypeOf((*SCCreateRoleAck)(nil)).Elem())
	log.Println("pbbind_gen.go init success")
}