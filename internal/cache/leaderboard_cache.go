package cache

import (
	"context"
	"strconv"

	"github.com/redis/go-redis/v9"
)

type LeaderboardCacheService struct {
	cache CacheService
}

func NewLeaderboardCacheService(cache CacheService) *LeaderboardCacheService {
	return &LeaderboardCacheService{cache: cache}
}

// 更新用户排行榜分数
func (l *LeaderboardCacheService) UpdateUserScore(ctx context.Context, leaderboardType string, userID uint64, score float64) error {
	leaderboardKey := LeaderboardKey(leaderboardType)
	return l.cache.ZAdd(ctx, leaderboardKey, redis.Z{
		Score:  score,
		Member: userID,
	})
}

// 获取排行榜前N名
func (l *LeaderboardCacheService) GetTopN(ctx context.Context, leaderboardType string, n int) ([]redis.Z, error) {
	leaderboardKey := LeaderboardKey(leaderboardType)
	scores, err := l.cache.ZRangeWithScores(ctx, leaderboardKey, 0, int64(n-1))
	if err != nil {
		return nil, err
	}
	return scores, nil
}

// 获取用户排名
func (l *LeaderboardCacheService) GetUserRank(ctx context.Context, leaderboardType string, userID uint64) (int64, error) {
	leaderboardKey := LeaderboardKey(leaderboardType)
	rank, err := l.cache.ZRevRank(ctx, leaderboardKey, strconv.FormatUint(userID, 10))
	if err != nil {
		return 0, err
	}
	return rank + 1, nil
}

// 获取用户分数
func (l *LeaderboardCacheService) GetUserScore(ctx context.Context, leaderboardType string, userID uint64) (float64, error) {
	leaderboardKey := LeaderboardKey(leaderboardType)
	score, err := l.cache.ZScore(ctx, leaderboardKey, strconv.FormatUint(userID, 10))
	if err != nil {
		return 0, err
	}
	return score, nil
}

// 从排行榜移除用户
func (l *LeaderboardCacheService) RemoveUserFromLeaderboard(ctx context.Context, leaderboardType string, userID uint64) error {
	leaderboardKey := LeaderboardKey(leaderboardType)
	return l.cache.ZRem(ctx, leaderboardKey, userID)
}
