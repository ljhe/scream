package mpool

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestMemoryPool(t *testing.T) {
	MemoryPoolInit()
	pools := GetMemoryPool(SystemMemoryPoolKey)
	for i := 0; i < 20; i++ {
		go func() {
			mem := pools.Get(7)
			fmt.Printf("test memory pool. i:%d  size:%d 地址:%p \n", i, len(mem), mem)
			pools.Put(mem)
		}()
	}
	time.Sleep(3 * time.Second)
	cur, max := pools.GetCount(8)
	fmt.Printf("test memory pool end. cur:%d max:%d \n", cur, max)
	for i := 0; i < 20; i++ {
		go func() {
			mem := pools.Get(7)
			fmt.Printf("test memory pool after sleep. i:%d  size:%d 地址:%p \n", i, len(mem), mem)
			pools.Put(mem)
		}()
	}
	time.Sleep(2 * time.Second)
}

// 只有在高并
// cpu: 12th Gen Intel(R) Core(TM) i7-12700
// j = 100000	size = 40960
// BenchmarkMemoryPool-20    	  295358	      4108 ns/op
// j = 100000	size = 40960
// BenchmarkNotMemoryPool-20      1000000	      1400 ns/op
// j = 1000000	size = 40960
// BenchmarkMemoryPool-20    	  299025	      3859 ns/op
// j = 1000000	size = 40960
// BenchmarkNotMemoryPool-20      130359	      11389 ns/op
// j = 1000000	size = 1024
// BenchmarkMemoryPool-20    	  299025	      3859 ns/op
// j = 1000000	size = 1024
// BenchmarkNotMemoryPool-20      147814	      9189 ns/op
// j = 10000000	size = 40960
// BenchmarkMemoryPool-20    	  275229	      4061 ns/op
// j = 10000000	size = 40960
// BenchmarkNotMemoryPool-20      18127	    	  124465 ns/op
func BenchmarkMemoryPool(b *testing.B) {
	MemoryPoolInit()
	pools := GetMemoryPool(SystemMemoryPoolKey)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		go func() {
			for j := 0; j < 1000000; j++ {
				mem := pools.Get(1024)
				mem[0] = 1
				pools.Put(mem)
			}
		}()
	}
}

func BenchmarkNotMemoryPool(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		go func() {
			for j := 0; j < 1000000; j++ {
				data := make([]byte, 40960)
				data[0] = 1
			}
		}()
	}
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
