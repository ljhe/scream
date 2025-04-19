package mpool

import (
	"fmt"
	"sync"
	"testing"
)

func TestMemoryPool(t *testing.T) {
	MemoryPoolInit()
	pools := GetMemoryPool(TCPMemoryPoolKey)
	cur, max := pools.GetCount(8)
	fmt.Printf("test memory pool begin. cur:%d max:%d \n", cur, max)
	for i := 0; i < 100; i++ {
		go func() {
			mem := pools.Get(7)
			fmt.Printf("test memory pool. i:%d  size:%d 地址:%p \n", i, len(mem), mem)
			pools.Put(mem)
		}()

		go func() {
			mem := pools.Get(7)
			fmt.Printf("test memory pool. i:%d  size:%d 地址:%p \n", i, len(mem), mem)
			pools.Put(mem)
		}()
	}
	cur, max = pools.GetCount(8)
	fmt.Printf("test memory pool end. cur:%d max:%d \n", cur, max)
}

type TMemoryPool struct {
	pool sync.Pool
}

func TNewMemoryPool() *TMemoryPool {
	return &TMemoryPool{
		pool: sync.Pool{
			New: func() interface{} {
				return make([]byte, 1024) // 假设每个内存块大小为 1024 字节
			},
		},
	}
}

func (mp *TMemoryPool) Get() []byte {
	return mp.pool.Get().([]byte)
}

func (mp *TMemoryPool) Put(b []byte) {
	mp.pool.Put(b)
}

func TestTNewMemoryPool(t *testing.T) {
	mp := TNewMemoryPool()

	// 测试内存池是否复用内存块
	for i := 0; i < 5; i++ {
		mem := mp.Get()
		fmt.Printf("获取的内存块地址：%p 大小:%d \n", mem, len(mem))
		mp.Put(mem) // 归还内存块
	}
}
