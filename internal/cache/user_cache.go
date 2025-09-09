package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/mangooer/gamehub-arena/internal/models"
)

type UserCacheService struct {
	cache CacheService
}

func NewUserCacheService(cache CacheService) *UserCacheService {
	return &UserCacheService{cache: cache}
}

// 用户会话
func (s *UserCacheService) SetUserSession(ctx context.Context, userID uint64, sessionData map[string]interface{}) error {
	sessionKey := UserSessionKey(userID)
	return s.cache.HSet(ctx, sessionKey, sessionData)
}

func (s *UserCacheService) GetUserSession(ctx context.Context, userID uint64) (map[string]string, error) {
	sessionKey := UserSessionKey(userID)
	return s.cache.HGetAll(ctx, sessionKey)
}

func (s *UserCacheService) DeleteUserSession(ctx context.Context, userID uint64) error {
	sessionKey := UserSessionKey(userID)
	return s.cache.Del(ctx, sessionKey)
}

// 在线用户管理
func (s *UserCacheService) AddOnlineUser(ctx context.Context, userID uint64) error {
	return s.cache.SAdd(ctx, UsersOnlineKey(), userID)
}

func (s *UserCacheService) RemoveOnlineUser(ctx context.Context, userID uint64) error {
	return s.cache.SRem(ctx, UsersOnlineKey(), userID)
}

func (s *UserCacheService) GetOnlineUsers(ctx context.Context) ([]string, error) {
	return s.cache.SMembers(ctx, UsersOnlineKey())
}

func (s *UserCacheService) GetOnlineUserCount(ctx context.Context) (int64, error) {
	return s.cache.SCard(ctx, UsersOnlineKey())
}

// 用户信息缓存
func (s *UserCacheService) SetUserInfo(ctx context.Context, user *models.User, expiration time.Duration) error {
	userKey := UserInfoKey(user.ID)
	userJsonData, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return s.cache.Set(ctx, userKey, userJsonData, expiration)
}

// 获取用户信息
func (s *UserCacheService) GetUserInfo(ctx context.Context, userID uint64) (*models.User, error) {
	userKey := UserInfoKey(userID)
	userJsonData, err := s.cache.Get(ctx, userKey)
	if err != nil {
		return nil, err
	}
	var user models.User
	if err := json.Unmarshal([]byte(userJsonData), &user); err != nil {
		return nil, err
	}

	return &user, nil

}

// 用户状态管理
func (s *UserCacheService) SetUserStatus(ctx context.Context, userID uint64, status string) error {
	statusKey := UserStatusKey(userID)
	return s.cache.Set(ctx, statusKey, status, 24*time.Hour)
}

func (s *UserCacheService) GetUserStatus(ctx context.Context, userID uint64) (string, error) {
	statusKey := UserStatusKey(userID)
	return s.cache.Get(ctx, statusKey)
}
