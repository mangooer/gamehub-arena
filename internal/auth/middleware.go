package auth

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type contextKey string

const (
	UserContextKey contextKey = "user"
)

// 用户上下文信息
type UserContext struct {
	UserID   uint64
	Username string
	Email    string
	Roles    []string
}

type AuthMiddleware struct {
	jwtService *JWTService
}

func NewAuthMiddleware(jwtService *JWTService) *AuthMiddleware {
	return &AuthMiddleware{jwtService: jwtService}
}

func (a *AuthMiddleware) Authenticate(ctx context.Context) (*UserContext, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "metadata is required")
	}
	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing authorization token")
	}

	authHeader := authHeaders[0]
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return nil, status.Error(codes.Unauthenticated, "invalid authorization token")
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := a.jwtService.ValidateToken(token)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}
	userContext := &UserContext{
		UserID:   claims.UserID,
		Username: claims.Username,
		Email:    claims.Email,
		Roles:    claims.Roles,
	}
	return userContext, nil
}

// 是否是公共方法
func (a *AuthMiddleware) IsPublicMethod(method string) bool {
	publicMethods := []string{
		"/health.v1.HealthService/Check",
		"/health.v1.HealthService/Watch",
	}
	for _, publicMethod := range publicMethods {
		if method == publicMethod {
			return true
		}
	}
	return false
}

// grpc拦截器
func (a *AuthMiddleware) UnaryAuthInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	if a.IsPublicMethod(info.FullMethod) {
		return handler(ctx, req)
	}
	userContext, err := a.Authenticate(ctx)
	if err != nil {
		return nil, err
	}
	// 将用户信息添加到上下文
	newCtx := context.WithValue(ctx, UserContextKey, userContext)
	return handler(newCtx, req)
}

// stream拦截器
func (a *AuthMiddleware) StreamAuthInterceptor(
	srv interface{},
	stream grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	if a.IsPublicMethod(info.FullMethod) {
		return handler(srv, stream)
	}
	userContext, err := a.Authenticate(stream.Context())
	if err != nil {
		return err
	}
	// 将用户信息添加到上下文
	newCtx := context.WithValue(stream.Context(), UserContextKey, userContext)
	wrappedStream := &wrappedServerStream{
		ServerStream: stream,
		context:      newCtx,
	}
	return handler(srv, wrappedStream)
}

// wrappedServerStream 包装ServerStream
type wrappedServerStream struct {
	grpc.ServerStream
	context context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.context
}

// 授权检查
func (a *AuthMiddleware) RequireRole(requiredRoles string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		userContext, ok := ctx.Value(UserContextKey).(*UserContext)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "user context not found")
		}
		hasRole := false
		for _, role := range userContext.Roles {
			if role == requiredRoles {
				hasRole = true
				break
			}
		}
		if !hasRole {
			return nil, status.Error(codes.PermissionDenied, "user does not have the required role")
		}
		return handler(ctx, req)
	}
}

// 获取当前用户上下文
func (a *AuthMiddleware) GetUserContext(ctx context.Context) (*UserContext, error) {
	// 获取当前用户上下文并断言
	userContext, ok := ctx.Value(UserContextKey).(*UserContext)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user context not found")
	}
	return userContext, nil
}
