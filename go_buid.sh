cur_path=`pwd`
export GOPATH=$cur_path:$GOPATH
go build -o bin/commitLog src/main.go
go build -o bin/client src/rpc_client/client.go
