#!/bin/bash

# 设置变量
PROTO_DIR="api/proto"
GO_OUT_DIR="api/gen/go"

# 检查依赖
if ! command -v protoc &> /dev/null; then
    echo "Error: protoc is not installed"
    exit 1
fi

if ! command -v protoc-gen-go &> /dev/null; then
    echo "Error: protoc-gen-go is not installed"
    echo "Run: go install google.golang.org/protobuf/cmd/protoc-gen-go@latest"
    exit 1
fi

if ! command -v protoc-gen-go-grpc &> /dev/null; then
    echo "Error: protoc-gen-go-grpc is not installed"
    echo "Run: go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest"
    exit 1
fi

# 清理生成目录
if [ -d "$GO_OUT_DIR" ]; then
    rm -rf "$GO_OUT_DIR"
fi

# 创建生成目录
mkdir -p "$GO_OUT_DIR"

echo "Generating protobuf code..."

# 遍历proto目录下的所有proto文件
for proto_file in $(find $PROTO_DIR -name "*.proto"); do
    proto_name=$(basename $proto_file)
    echo "Generating: $proto_name"
    
    protoc \
        --proto_path="$PROTO_DIR" \
        --go_out="$GO_OUT_DIR" \
        --go_opt=paths=source_relative \
        --go-grpc_out="$GO_OUT_DIR" \
        --go-grpc_opt=paths=source_relative \
        "$proto_file"
    
    if [ $? -ne 0 ]; then
        echo "Error generating $proto_name"
        exit 1
    fi
done

echo "Proto files generated successfully!"