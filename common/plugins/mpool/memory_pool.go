package mpool

import (
	"log"
	"sync"
)

type MemoryPoolItem struct {
	pool     sync.Pool
	maxCount int // 最大缓存数量
	curCount int // 当前缓存数量
	mu       sync.Mutex
}

type MemoryPool struct {
	pools map[int]*MemoryPoolItem
	sizes []int
}

func NewMemoryPool(mps, mpc []int) *MemoryPool {
	pools := make(map[int]*MemoryPoolItem)
	for k, size := range mps {
		pools[size] = &MemoryPoolItem{
			pool: sync.Pool{
				New: func() interface{} {
					return make([]byte, size)
				},
			},
			maxCount: mpc[k],
		}
		log.Printf("Allocating new memory of size: %d maxCount:%d \n", size, mpc[k])
	}
	return &MemoryPool{
		pools: pools,
		sizes: mps,
	}
}

// 找到大于或等于请求大小的最接近的块大小
func (mp *MemoryPool) findClosestSize(size int) int {
	for _, s := range mp.sizes {
		if size <= s {
			return s
		}
	}
	return mp.sizes[len(mp.sizes)-1]
}

// Get 获取内存块
func (mp *MemoryPool) Get(size int) []byte {
	closestSize := mp.findClosestSize(size)
	return mp.pools[closestSize].pool.Get().([]byte)
}

// Put 将内存块放回池中
func (mp *MemoryPool) Put(buf []byte) {
	closestSize := mp.findClosestSize(cap(buf))
	item := mp.pools[closestSize]

	item.mu.Lock()
	defer item.mu.Unlock()

	// 只有在缓存数量小于最大限制时才将块放回池中
	if item.curCount < item.maxCount {
		item.curCount++
		item.pool.Put(buf)
	} else {
		// 超过最大缓存数量的块直接丢弃
		//log.Printf("Discarding memory block of size: %d, as pool is full\n", closestSize)
	}
}
