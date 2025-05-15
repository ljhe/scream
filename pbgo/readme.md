#protoc.exe proto-gen-go.exe 用来生成protobuf文件
#编译某一个pb文件
#.\protoc --proto_path=../protobuf --go_out=./ login.proto

#编译单个*.go文件
    进入该目录下
    go build -o .\pbgo\proto_cmd .\pbgo\proto_cmd\proto_cmd.go