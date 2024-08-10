package common

type ServerNodeProperty interface {
	GetAddr() string
	SetAddr(s string)
}
