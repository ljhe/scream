@echo off
echo "building proto files"

:: protobuf源文件的路径
set protoPath="..\protobuf"

echo "copying proto files"
xcopy /s /y %protoPath% .
echo "copying proto files success"

"%~dp0proto_cmd/proto_cmd.exe"

for %%i in (*.proto) do (
    protoc --go_out=. "%%i"
)

if exist "messagedef.proto" (
    xcopy /s /y "messagedef.proto" %protoPath%
)

echo "deleting proto files"
del *.proto

echo "building proto files success"
pause