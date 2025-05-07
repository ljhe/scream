package mongo

import "time"

var config = struct {
	connectTimeout time.Duration
	MaxPoolSize    uint64
}{
	connectTimeout: 5 * time.Second,
	MaxPoolSize:    50,
}
