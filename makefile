.PHONY: build clean test run proto install-proto

# 构建所有服务
build:
	go build -o bin/gateway cmd/gateway/main.go
	go build -o bin/match-service cmd/match-service/main.go
	go build -o bin/room-service cmd/room-service/main.go

# 清理构建文件
clean:
	@rm -rf bin/
	@rm -rf api/gen/

# 运行测试
test:
	go test ./...

# 运行网关服务
run-gateway:
	go run cmd/gateway/main.go

# 格式化代码
fmt:
	go fmt ./...

# 代码检查
lint:
	golangci-lint run

# 生成protobuf文件 (跨平台)
proto:
	@echo "Generating proto files (Unix)"
	@chmod +x scripts/generate-proto.sh
	@./scripts/generate-proto.sh

# 安装protobuf工具 (跨平台)
install-proto:
	@echo "Installing protobuf tools"
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@echo "Protobuf tools installed successfully!"
	@echo "Please ensure protoc is installed: https://github.com/protocolbuffers/protobuf/releases"
