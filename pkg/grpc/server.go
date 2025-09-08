package grpc

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	server *grpc.Server
	port   int
}

func NewServer(port int, opts ...grpc.ServerOption) *Server {

	//默认选项
	defaultOpts := []grpc.ServerOption{
		grpc.UnaryInterceptor(unaryInterceptor),
		grpc.StreamInterceptor(streamInterceptor),
	}

	opts = append(defaultOpts, opts...)

	server := grpc.NewServer(opts...)

	// 注册健康服务检查
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(server, healthServer)

	// 注册反射服务
	reflection.Register(server)

	return &Server{
		server: server,
		port:   port,
	}
}

// RegisterService 注册服务
func (s *Server) RegisterService(registerFunc func(*grpc.Server)) {
	registerFunc(s.server)
}

// Start 启动服务
func (s *Server) Start() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}
	fmt.Printf("gRPC server listening on port %d", s.port)
	return s.server.Serve(lis)
}

// Stop 停止服务
func (s *Server) Stop() {
	s.server.GracefulStop()
}

// unaryInterceptor 拦截器
func unaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	return handler(ctx, req)
}

// streamInterceptor 拦截器
func streamInterceptor(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	return handler(srv, stream)
}
