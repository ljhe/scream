package main

import (
	"common"
	"common/service"
	_ "common/socket/tcp"
	"log"
)

func main() {
	log.Println("server starting ...")
	service.Init()
	node := service.CreateAcceptor(service.NetNodeParam{
		ServerTyp:  common.SocketTypTcpAcceptor,
		ServerName: "test",
		Addr:       "0.0.0.0:2701",
		Typ:        1,
		Zone:       9999,
		Index:      1,
	})
	log.Println("server start success")
	service.WaitExitSignal()
	log.Println("server stopping ...")
	service.Stop(node)
	log.Println("server close")
}
