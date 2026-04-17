BINARY_NAME=deploy-agent
build: 
	GOARCH = amd64 GOOS = linux go build -o bin/${BINARY_NAME}-linux cmd/agent/main.go
	go build -o bin/${BINARY_NAME}-local cmd/agent/main.go

run:
	go run cmd/agent/main.go

clean: 
	go clean
	rm -rf bin/
	rm -rf ./tmp-build-*

tidy:
	go mod tidy