cur_path=`pwd`
export GOPATH=$cur_path:$GOPATH
go build -o bin/commitLog src/*.go
