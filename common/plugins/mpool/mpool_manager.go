package mpool

import (
	"common"
)

var TCPMemoryPoolKey = "TCPMemoryPoolKey"

var MemoryPoolConfigs = []*MemoryPoolConf{
	{
		key: TCPMemoryPoolKey,
		Mps: []int{32, 64, 128, 256, 512, 1024, 2048, 4096, 8192, 16384, 32768, common.MsgMaxLen},
		Mpc: []int{4, 4, 4, 4, 4, 4, 4, 2, 2, 2, 2, 2},
	},
}

var MemoryPoolObj = map[string]*MemoryPool{}

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

func NewMemoryPoolManager(mpc *MemoryPoolConf) *MemoryPool {
	return NewMemoryPool(mpc.Mps, mpc.Mpc)
}

func GetMemoryPool(key string) *MemoryPool {
	return MemoryPoolObj[key]
}
