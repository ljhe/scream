package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

var ServerPreKey = "server/"

type ServerInfo struct {
	Id      string
	Name    string
	Host    string
	Typ     int
	Index   int
	RegTime int64
}

func (e *ServerInfo) String() string {
	data, err := json.Marshal(e)
	if err != nil {
		return ""
	}
	return string(data)
}

// GenSelfServiceId 生成自己的服务器id
func GenSelfServiceId(name string, typ, index int) string {
	return fmt.Sprintf("%s#%d@%d",
		name,
		typ,
		index,
	)
}

// GenTargetServiceId 根据对方的服务器名称 和自己的zone type index 来获取对方的服务器key
// 例如某一组服务器中 两个相对应的节点
func GenTargetServiceId(target string, typ, index int) string {
	return fmt.Sprintf("%s#%d@%d",
		target,
		typ,
		index,
	)
}

func GenServicePrefix(name string) string {
	return ServerPreKey + name
}

func GenDiscoveryServicePrefix(name string) string {
	return ServerPreKey + name + "#"
}

func ParseServiceId(sid string) (typ, idx int, err error) {
	str := strings.Split(sid, "#")
	if len(str) < 2 {
		err = errors.New(fmt.Sprintf("ParseServiceId sid invalid. sid:%s", sid))
		return
	} else {
		strProp := strings.Split(str[1], "@")
		if len(strProp) < 3 {
			err = errors.New(fmt.Sprintf("ParseServiceId sid invalid. sid:%s", sid))
			return
		} else {
			typ, err = StrToInt(strProp[0])
			if err != nil {
				return
			}
			idx, err = StrToInt(strProp[1])
			if err != nil {
				return
			}
		}
	}
	return
}

func WaitExitSignal() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL)
	<-ch
}
