module pbgo

go 1.22.4

require (
	common v0.0.0-00010101000000-000000000000
	github.com/golang/protobuf v1.5.4
)

require google.golang.org/protobuf v1.34.1 // indirect

replace common => ../common
