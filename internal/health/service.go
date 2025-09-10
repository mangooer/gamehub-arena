package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/mangooer/gamehub-arena/api/gen/go/common"
	health_v1 "github.com/mangooer/gamehub-arena/api/gen/go/health/v1"
	"github.com/mangooer/gamehub-arena/internal/cache"
	"github.com/mangooer/gamehub-arena/internal/database"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Service struct {
	health_v1.UnimplementedHealthServiceServer
	database  *database.Database
	cache     *cache.RedisClient
	startTime time.Time
}

func NewService(database *database.Database, cache *cache.RedisClient) *Service {
	return &Service{
		database:  database,
		cache:     cache,
		startTime: time.Now(),
	}
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

func (s *Service) DetailedCheck(ctx context.Context, req *health_v1.DetailedHealthCheckRequest) (*health_v1.HealthCheckResponse, error) {
	services := make(map[string]string)
	overAllStatus := health_v1.HealthCheckResponse_SERVING
	message := "All services are running normally"
	if s.database != nil {
		if err := s.database.Ping(); err != nil {
			services["database"] = "not serving"
			overAllStatus = health_v1.HealthCheckResponse_NOT_SERVING
			message = "Database connection failed"
		} else {
			services["database"] = "serving"
		}
	} else {
		services["database"] = "not available"
		overAllStatus = health_v1.HealthCheckResponse_NOT_SERVING
		message = "Database connection not available"
	}
	if s.cache != nil {
		if err := s.cache.Ping(ctx); err != nil {
			services["cache"] = "not serving"
			overAllStatus = health_v1.HealthCheckResponse_NOT_SERVING
			message = "Cache connection failed"
		} else {
			services["cache"] = "serving"
		}
	} else {
		services["cache"] = "not available"
		overAllStatus = health_v1.HealthCheckResponse_NOT_SERVING
		message = "Cache connection not available"
	}

	return &health_v1.HealthCheckResponse{
		Status:        overAllStatus,
		Message:       message,
		Services:      services,
		UptimeSeconds: int64(time.Since(s.startTime).Seconds()),
	}, nil

}

func (s *Service) handleHTTPHealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	resp, err := s.Check(ctx, &health_v1.HealthCheckRequest{})
	if err != nil {

		// 使用BaseResponse结构，包含详细信息
		baseResp := &common.BaseResponse{
			Code:      200,
			Message:   resp.Message,
			Timestamp: timestamppb.Now(),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(baseResp)
		return
	}

	// 使用BaseResponse结构
	baseResp := &common.BaseResponse{
		Code:      200,
		Message:   "Health check passed",
		Timestamp: timestamppb.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(baseResp)
}

func (s *Service) handleDetailedHTTPHealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	resp, err := s.DetailedCheck(ctx, &health_v1.DetailedHealthCheckRequest{
		IncludeDependencies: true,
	})
	if err != nil {
		baseResp := &common.BaseResponse{
			Code:      200,
			Message:   resp.Message,
			Timestamp: timestamppb.Now(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(baseResp)
		return
	}

	// 使用BaseResponse结构，包含详细信息
	baseResp := &common.BaseResponse{
		Code:      200,
		Message:   resp.Message,
		Timestamp: timestamppb.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(baseResp)
}
