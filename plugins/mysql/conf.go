package mysql

var GormConfig = struct {
	MaxIdleConns int
	MaxOpenConns int
}{
	MaxIdleConns: 32,
	MaxOpenConns: 100,
}
