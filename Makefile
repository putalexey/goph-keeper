# These are the values we want to pass for VERSION and BUILD
# git tag 1.0.1
# git commit -am "One more change after the tags"
VERSION=`git describe --tags`
BUILD=`date +%FT%T%z`
#LDFLAGS=-ldflags "-w -s -X main.buildVersion=${VERSION} -X main.buildDate=${BUILD}"
LDFLAGS=-ldflags "-X main.buildVersion=${VERSION} -X main.buildDate=${BUILD}"

#proto:
#	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative internal/app/proto/grpc.proto

build: build_client build_server

build_client:
	go build ${LDFLAGS} -o bin/gk-client ./cmd/gk-client

build_server:
	go build ${LDFLAGS} -o bin/gk-server ./cmd/gk-server

proto:
	cd internal/common/gproto; protoc -I=. --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ./models.proto ./services.proto

#docs:
#	cd internal/app/shortener/
#	swag init -g ./shortner.go
