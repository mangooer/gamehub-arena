package health

import (
	"context"

	health_v1 "github.com/mangooer/gamehub-arena/api/gen/go/health/v1"
)

type Service struct {
	health_v1.UnimplementedHealthServiceServer
}

func NewService() *Service {
	return &Service{}
}

// Check 检查服务健康状态
func (s *Service) Check(ctx context.Context, req *health_v1.HealthCheckRequest) (*health_v1.HealthCheckResponse, error) {
	return &health_v1.HealthCheckResponse{
		Status: health_v1.HealthCheckResponse_SERVING,
	}, nil
}

// Watch 监控服务状态变化
func (s *Service) Watch(req *health_v1.HealthCheckRequest, stream health_v1.HealthService_WatchServer) error {
	// 实现状态监控逻辑
	return stream.Send(&health_v1.HealthCheckResponse{
		Status: health_v1.HealthCheckResponse_SERVING,
	})
}
