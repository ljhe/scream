package utils

import (
	"fmt"
	"github.com/ljhe/scream/core/iface"
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
