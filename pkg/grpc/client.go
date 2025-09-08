package grpc

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ClientConfig struct {
	Address string
	Timeout time.Duration
}

func NewClient(config ClientConfig) (*grpc.ClientConn, error) {
	// 使用新的 grpc.NewClient API
	conn, err := grpc.NewClient(config.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create client for %s: %v", config.Address, err)
	}

	// 测试连接（可选）
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	// 可以通过健康检查等方式验证连接
	_ = ctx // 如果需要连接验证，可以在这里添加逻辑

	return conn, nil
}
