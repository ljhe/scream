package utils

import (
	"fmt"
	"github.com/ljhe/scream/core/iface"
	"os"
	"os/signal"
	"syscall"
)

// GenServiceId 生成服务器id
func GenServiceId(prop iface.INodeProp) string {
	return fmt.Sprintf("%s#%d@%d@%d",
		prop.GetName(),
		prop.GetZone(),
		prop.GetServerTyp(),
		prop.GetIndex(),
	)
}

func WaitExitSignal() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL)
	<-ch
}
