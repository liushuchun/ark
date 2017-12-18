DIR=$(cd ../; pwd)
export GOPATH=$DIR:$GOPATH
go build -o fuck main.go
