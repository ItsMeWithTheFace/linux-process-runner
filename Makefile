.PHONY: proto logs certs build

build-all: proto logs certs build

certs:
	cd certs && ./gen.sh

proto:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative api/proto/api.proto

logs:
	sudo mkdir /var/log/linux-process-runner

build: certs logs
	go build -o bin/server main/server/server.go
	go build -o bin/client main/client/client.go

tests:
	sudo go test ./... -v
