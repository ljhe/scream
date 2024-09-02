package util

import (
	"common"
	"fmt"
)

// GenServiceId 生成服务器id
func GenServiceId(prop common.ServerNodeProperty) string {
	return fmt.Sprintf("%s#%d@%d@%d",
		prop.GetName(),
		prop.GetZone(),
		prop.GetServerTyp(),
		prop.GetIndex(),
	)
}
