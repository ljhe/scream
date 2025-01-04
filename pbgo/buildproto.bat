@echo off
echo "building proto files"

:: protobuf源文件的路径
set protoPath="..\protobuf\*.proto"

echo "copying proto files"
xcopy /s /y %protoPath% .
echo "copying proto files success"

for %%i in (*.proto) do (
    protoc --go_out=. "%%i"
)

echo "deleting proto files"
del *.proto

echo "building proto files success"