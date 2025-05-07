module pbgo

go 1.24.1

require (
	common v0.0.0-00010101000000-000000000000
	github.com/golang/protobuf v1.5.4
)

require google.golang.org/protobuf v1.36.5 // indirect

replace common => ../common
