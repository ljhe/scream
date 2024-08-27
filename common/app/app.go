package app

import (
	"common/iface"
	plugins "common/plugins/etcd"
	_ "common/socket/tcp"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func Init() {
	err := plugins.InitServiceDiscovery("127.0.0.1:2379")
	if err != nil {
		log.Println("InitServiceDiscovery err:", err)
		return
	}
}

func WaitExitSignal() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL)
	<-ch
}

func Stop(node iface.INetNode) {
	if node == nil {
		return
	}
	node.Stop()
}
