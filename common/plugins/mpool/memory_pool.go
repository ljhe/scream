package mpool

import (
	"fmt"
	"log"
	"sync"
	"sync/atomic"
)

type MemoryPool struct {
	pool sync.Pool
	max  int32 // 最大缓存数量
	cur  int32 // 当前缓存数量
	mu   sync.Mutex
}

type MemoryPools struct {
	pools map[int]*MemoryPool
	sizes []int
}

func NewMemoryPools(mps, mpc []int) *MemoryPools {
	pools := make(map[int]*MemoryPool)
	for k, size := range mps {
		pool := &MemoryPool{
			pool: sync.Pool{
				New: func() interface{} {
					return make([]byte, size)
				},
			},
			max: int32(mpc[k]),
		}
		atomic.AddInt32(&pool.cur, 1)
		pools[size] = pool
		log.Printf("Allocating new memory of size: %d maxCount:%d \n", size, mpc[k])
	}
	return &MemoryPools{
		pools: pools,
		sizes: mps,
	}
}

// 找到大于或等于请求大小的最接近的块大小
func (mps *MemoryPools) findClosestSize(size int) int {
	for _, s := range mps.sizes {
		if size <= s {
			return s
		}
	}
	return mps.sizes[len(mps.sizes)-1]
}

// Get 获取内存块
func (mps *MemoryPools) Get(size int) []byte {
	closestSize := mps.findClosestSize(size)
	pool := mps.pools[closestSize]
	if pool == nil {
		panic(fmt.Sprintf("memory pool is nil. size:%d closestSize:%d", size, closestSize))
	}
	atomic.AddInt32(&pool.cur, -1)
	return pool.pool.Get().([]byte)
}

// Put 将内存块放回池中
func (mps *MemoryPools) Put(buf []byte) {
	closestSize := mps.findClosestSize(len(buf))
	pool := mps.pools[closestSize]

	pool.mu.Lock()
	defer pool.mu.Unlock()

	// 只有在缓存数量小于最大限制时才将块放回池中
	if pool.cur < pool.max {
		atomic.AddInt32(&pool.cur, 1)
		pool.pool.Put(buf)
	} else {
		// 超过最大缓存数量的块直接丢弃
		//log.Printf("Discarding memory block of size: %d, as pool is full\n", closestSize)
	}
}

func (mps *MemoryPools) GetCount(size int) (int32, int32) {
	pool := mps.pools[size]
	if pool == nil {
		return 0, 0
	}
	return pool.cur, pool.max
}

func (mp *MemoryPool) Put(b []byte) {
	mp.pool.Put(b)
}
