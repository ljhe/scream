package util

import (
	"fmt"
	"github.com/ljhe/scream/common/iface"
)

// GenServiceId 生成服务器id
func GenServiceId(prop iface.ServerNodeProperty) string {
	return fmt.Sprintf("%s#%d@%d@%d",
		prop.GetName(),
		prop.GetZone(),
		prop.GetServerTyp(),
		prop.GetIndex(),
	)
}
