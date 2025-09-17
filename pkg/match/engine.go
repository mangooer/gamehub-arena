package match

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/mangooer/gamehub-arena/internal/cache"
	"github.com/mangooer/gamehub-arena/internal/config"
	"github.com/mangooer/gamehub-arena/internal/logger"
	"github.com/mangooer/gamehub-arena/pkg/algorithm"
	"go.uber.org/zap"
)

type MatchingEngine struct {
	algorithm    algorithm.MatchingAlgorithm
	queueManager *QueueManager
	cache        cache.CacheService
	config       *config.Config
	logger       logger.Logger

	// 运行控制
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// 统计信息
	stats   *EngineStats
	statsMu sync.RWMutex
}

type EngineStats struct {
	TotalRequests     int64         `json:"total_requests"`
	SuccessfulMatches int64         `json:"successful_matches"`
	FailedMatches     int64         `json:"failed_matches"`
	AverageMatchTime  time.Duration `json:"average_match_time"`
	QueueSize         int           `json:"queue_size"`
	ActiveWorkers     int           `json:"active_workers"`
	LastUpdated       time.Time     `json:"last_updated"`
}

type MMRRange struct {
	Name   string  `json:"name"`
	MinMMR float64 `json:"min_mmr"`
	MaxMMR float64 `json:"max_mmr"`
}

func NewMatchingEngine(algorithmName string, queueManager *QueueManager, cache cache.CacheService, config *config.Config, logger logger.Logger) (*MatchingEngine, error) {
	factory := algorithm.InitFactory()
	algorithm, err := factory.GetAlgorithm(algorithmName)
	if err != nil {
		logger.GetLogger().Error("failed to get algorithm",
			zap.String("algorithm_name", algorithmName),
			zap.Error(err),
		)
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &MatchingEngine{
		algorithm:    algorithm,
		queueManager: queueManager,
		cache:        cache,
		config:       config,
		logger:       logger,
		ctx:          ctx,
		cancel:       cancel,
		stats: &EngineStats{
			TotalRequests:     0,
			SuccessfulMatches: 0,
			FailedMatches:     0,
			AverageMatchTime:  0,
			QueueSize:         0,
			ActiveWorkers:     0,
			LastUpdated:       time.Now(),
		},
	}, nil
}

func (e *MatchingEngine) Start() error {
	e.logger.GetLogger().Info("Starting matching engine",
		zap.String("algorithm", e.algorithm.Name()),
		zap.String("version", e.algorithm.Version()),
	)
	// 启动多个携程处理不同等级的队列
	ranks := []MMRRange{
		{Name: "Beginner", MinMMR: 0, MaxMMR: 1000},
		{Name: "Bronze", MinMMR: 800, MaxMMR: 1200},       // 有重叠
		{Name: "Silver", MinMMR: 1100, MaxMMR: 1400},      // 有重叠
		{Name: "Gold", MinMMR: 1300, MaxMMR: 1600},        // 有重叠
		{Name: "Platinum", MinMMR: 1500, MaxMMR: 1800},    // 有重叠
		{Name: "Diamond", MinMMR: 1700, MaxMMR: 2100},     // 有重叠
		{Name: "Master", MinMMR: 2000, MaxMMR: 2500},      // 有重叠
		{Name: "Grandmaster", MinMMR: 2300, MaxMMR: 3000}, // 有重叠
	}

	for _, rank := range ranks {
		e.wg.Add(1)
		go e.processRankQueue(rank)
	}
	// 启动统计监控协程
	e.wg.Add(1)
	go e.monitorMatchingStats()
	// 启动定期清理超时匹配协程
	e.wg.Add(1)
	go e.cleanupExpiredMatches()
	return nil
}

func (e *MatchingEngine) Stop() error {
	e.logger.GetLogger().Info("Stopping matching engine")

	e.cancel()
	e.wg.Wait()

	e.logger.GetLogger().Info("Matching engine stopped")
	return nil
}

// 为玩家寻找匹配
func (e *MatchingEngine) FindMatch(ctx context.Context, player *algorithm.Player) (*algorithm.MatchResult, error) {
	startTime := time.Now()

	e.updateStats(func(stats *EngineStats) {
		stats.TotalRequests++
	})

	defer func() {
		duration := time.Since(startTime)
		e.updateStats(func(stats *EngineStats) {
			stats.AverageMatchTime = (stats.AverageMatchTime*time.Duration(stats.TotalRequests-1) + duration) / time.Duration(stats.TotalRequests)
		})
	}()

	// 获取候选玩家
	candidates, err := e.queueManager.GetCandidates(ctx, player)
	if err != nil {
		e.updateStats(func(stats *EngineStats) {
			stats.FailedMatches++
		})
		return nil, fmt.Errorf("failed to get candidates: %w", err)
	}

	// 寻找最佳匹配
	result, err := e.algorithm.FindOptimalMatch(ctx, player, candidates)
	if err != nil {
		e.updateStats(func(stats *EngineStats) {
			stats.FailedMatches++
		})
		return nil, fmt.Errorf("failed to find match: %w", err)
	}

	e.updateStats(func(stats *EngineStats) {
		stats.SuccessfulMatches++
	})

	e.logger.GetLogger().Info("Match found",
		zap.String("match_id", result.MatchID),
		zap.Float64("quality", result.Quality),
		zap.Int("players", len(result.Players)),
		zap.Duration("duration", time.Since(startTime)),
	)

	return result, nil

}

func (e *MatchingEngine) updateStats(fn func(*EngineStats)) {
	e.statsMu.Lock()
	defer e.statsMu.Unlock()
	fn(e.stats)
}

func (e *MatchingEngine) GetStats() *EngineStats {
	e.statsMu.RLock()
	defer e.statsMu.RUnlock()
	return e.stats
}

func (e *MatchingEngine) SwitchAlgorithm(algorithmName string) error {

	factory := algorithm.InitFactory()
	algorithm, err := factory.GetAlgorithm(algorithmName)
	if err != nil {
		return fmt.Errorf("failed to get algorithm: %w", err)
	}
	e.algorithm = algorithm
	e.logger.GetLogger().Info("Algorithm switched", zap.String("algorithm", algorithmName))
	return nil
}

func (e *MatchingEngine) processRankQueue(rank MMRRange) {
	defer e.wg.Done()
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-e.ctx.Done():
			return
		case <-ticker.C:
			e.processRankQueueOnce(rank)
		}
	}
}

func (e *MatchingEngine) processRankQueueOnce(rank MMRRange) {
	// 从队列获取等待匹配的玩家
	players, err := e.queueManager.GetWaitingPlayersByMMRRange(e.ctx, rank)
	if err != nil {
		e.logger.GetLogger().Error("failed to get waiting players",
			zap.String("rank", rank.Name),
			zap.Float64("min_mmr", rank.MinMMR),
			zap.Float64("max_mmr", rank.MaxMMR),
			zap.Error(err),
		)
		return
	}

	if len(players) < 2 {
		return
	}

	for _, player := range players {
		if _, err := e.FindMatch(e.ctx, player); err != nil {
			e.logger.GetLogger().Error("failed to find match",
				zap.Int64("playerID", player.ID),
				zap.String("rank", rank.Name),
				zap.Float64("min_mmr", rank.MinMMR),
				zap.Float64("max_mmr", rank.MaxMMR),
				zap.Error(err),
			)
		}
	}
}

func (e *MatchingEngine) monitorMatchingStats() {
	defer e.wg.Done()
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-e.ctx.Done():
			return
		case <-ticker.C:
			e.updateStats(func(stats *EngineStats) {
				stats.QueueSize = e.queueManager.GetTotalQueueSize()
				stats.LastUpdated = time.Now()
			})

			// 记录统计信息
			stats := e.GetStats()
			e.logger.GetLogger().Info("Engine stats",
				zap.Int64("total_requests", stats.TotalRequests),
				zap.Int64("successful_matches", stats.SuccessfulMatches),
				zap.Int64("failed_matches", stats.FailedMatches),
				zap.Int("queue_size", stats.QueueSize),
				zap.Duration("avg_match_time", stats.AverageMatchTime),
			)
		}
	}
}

func (e *MatchingEngine) cleanupExpiredMatches() {
	defer e.wg.Done()
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-e.ctx.Done():
			return
		case <-ticker.C:
			if err := e.queueManager.CleanupExpiredPlayers(e.ctx); err != nil {
				e.logger.GetLogger().Error("Failed to cleanup expired players",
					zap.Error(err),
				)
			}
		}
	}
}
