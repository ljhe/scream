package buffer

import "sync"

// A Pool is a type-safe wrapper around a sync.Pool.
type Pool struct {
	p *sync.Pool
}

// NewPool constructs a new Pool.
func NewPool() Pool {
	return Pool{p: &sync.Pool{
		New: func() interface{} {
			return &Buffer{bs: make([]byte, defaultSize, defaultSize)}
		},
	}}
}

// NewPoolWithDefaultSize constructs a new Pool with default size
func NewPoolWithDefaultSize(size int) Pool {
	return Pool{p: &sync.Pool{
		New: func() interface{} {
			return &Buffer{bs: make([]byte, size, size)}
		},
	}}
}

// Get retrieves a Buffer from the pool, creating one if necessary.
func (p Pool) Get() *Buffer {
	return p.GetWithSize(0)
}

// GetWithSize retrieves a Buffer with the provided size from the pool, creating one if necessary.
func (p Pool) GetWithSize(size int) *Buffer {
	buf := p.p.Get().(*Buffer)
	buf.Reset(size)
	buf.pool = p
	return buf
}

func (p Pool) put(buf *Buffer) {
	p.p.Put(buf)
}
