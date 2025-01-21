package pbgo

import (
	"log"
	"reflect"
)

func registerInfo(id uint16, msgType reflect.Type) {
	RegisterMessageInfo(&MessageInfo{ID: id, Codec: GetCodec(), Type: msgType})
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
	registerInfo(1002, reflect.TypeOf((*CSCreateRoleReq)(nil)).Elem())
	registerInfo(1003, reflect.TypeOf((*SCCreateRoleAck)(nil)).Elem())
	log.Println("pbbind_gen.go init success")
}