package mpool

import (
	"fmt"
	"runtime"
	"sync"
	"testing"
)

// go tests -bench . -benchmem

var pool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 0)
	},
}

func BenchmarkMake(b *testing.B) {
	for i := 0; i < b.N; i++ {
		// 每次分配新的切片
		_ = make([]byte, 40960)
	}
	printGCStats(b)
}

func BenchmarkSyncPool(b *testing.B) {
	for i := 0; i < b.N; i++ {
		obj := pool.Get().([]byte)
		// 模拟对对象的操作
		pool.Put(obj[:0])
	}
	printGCStats(b)
}

func BenchmarkSyncPoolWithoutPut(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = pool.Get().([]byte) // 只获取对象，不进行放回
	}
	printGCStats(b)
}

func BenchmarkMakeConcurrent(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// 每次分配新的切片
			_ = make([]byte, 40960)
		}
	})
	printGCStats(b)
}

func BenchmarkSyncPoolConcurrent(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			obj := pool.Get().([]byte)
			// 模拟对对象的操作
			pool.Put(obj[:0])
		}
	})
	printGCStats(b)
}

func BenchmarkSyncPoolWithoutPutConcurrent(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = pool.Get().([]byte) // 只获取对象，不进行放回
		}
	})
	printGCStats(b)
}

func printGCStats(b *testing.B) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	// 当前堆中已分配的内存字节数。此值表示当前应用程序使用的堆内存量
	b.Logf("HeapAlloc: %d bytes", memStats.HeapAlloc)
	// 堆中已分配的内存的总量（包括 GC 使用的内存）。它包括了 Go 程序分配的内存以及由垃圾回收器管理的内存。
	b.Logf("HeapSys: %d bytes", memStats.HeapSys)
	// 当前堆中未使用的内存字节数
	b.Logf("HeapIdle: %d bytes", memStats.HeapIdle)
	// 从应用程序中释放的堆内存字节数
	b.Logf("HeapReleased: %d bytes", memStats.HeapReleased)
	// 自程序启动以来发生的 GC 次数。通过这个值可以看出 GC 发生的频率
	b.Logf("NumGC: %d", memStats.NumGC)
	// 垃圾回收暂停总时间
	b.Logf("PauseTotalNs: %d", memStats.PauseTotalNs)
}

func TestSliceForAppend(t *testing.T) {
	for i := 10; i > 0; i-- {
		obj := pool.Get().([]byte)
		_, obj = sliceForAppend(obj[:0], i)
		fmt.Printf("pool's ptr: %p\n", obj)
		pool.Put(obj[:0])
	}
}

func sliceForAppend(in []byte, n int) (head, tail []byte) {
	if total := len(in) + n; cap(in) >= total {
		head = in[:total]
	} else {
		head = make([]byte, total)
		copy(head, in)
	}
	tail = head[len(in):]
	return
}
