   .PHONY: build clean test run
   
   # 构建所有服务
   build:
   	go build -o bin/gateway cmd/gateway/main.go
   	go build -o bin/match-service cmd/match-service/main.go
   	go build -o bin/room-service cmd/room-service/main.go
   
   # 清理构建文件
   clean:
   	rm -rf bin/
   
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