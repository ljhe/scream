module server

go 1.22.4

require common v0.0.1

require (
	github.com/panjf2000/ants/v2 v2.10.0 // indirect
	golang.org/x/sync v0.3.0 // indirect
)

replace (
	common => ../common
	github.com/coreos/bbolt => go.etcd.io/bbolt v1.3.5
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
)
