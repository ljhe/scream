package mpool

import (
	"github.com/ljhe/scream/common"
)

var SystemMemoryPoolKey = "SystemMemoryPoolKey"

var MemoryPoolConfigs = []*MemoryPoolConf{
	{
		key: SystemMemoryPoolKey,
		Mps: []int{8, 16, 32, 64, 128, 256, 512, 1024, 2048, 4096, 8192, 16384, 32768, common.MsgMaxLen},
		Mpc: []int{10, 10, 10, 10, 10, 10, 10, 10, 5, 5, 5, 5, 3, 3},
	},
}

var MemoryPoolObj = map[string]*MemoryPools{}

type MemoryPoolConf struct {
	Mps []int
	Mpc []int
	key string
}

func MemoryPoolInit() {
	for _, conf := range MemoryPoolConfigs {
		MemoryPoolObj[conf.key] = NewMemoryPoolManager(&MemoryPoolConf{
			Mps: conf.Mps,
			Mpc: conf.Mpc,
		})
	}
}

func NewMemoryPoolManager(mpc *MemoryPoolConf) *MemoryPools {
	return NewMemoryPools(mpc.Mps, mpc.Mpc)
}

func GetMemoryPool(key string) *MemoryPools {
	return MemoryPoolObj[key]
}
