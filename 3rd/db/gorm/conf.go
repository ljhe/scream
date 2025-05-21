package gorm

var config = struct {
	MaxIdleConns           int
	MaxOpenConns           int
	SkipDefaultTransaction bool // 禁用事务
}{
	MaxIdleConns:           32,
	MaxOpenConns:           100,
	SkipDefaultTransaction: true,
}
