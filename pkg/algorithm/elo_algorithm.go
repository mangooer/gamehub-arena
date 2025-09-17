package algorithm

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/mangooer/gamehub-arena/internal/config"
)

type ELOAlgorithm struct {
	config *config.AlgorithmConfig
	stats  *AlgorithmStats
	mu     sync.RWMutex
}

func NewELOAlgorithm(config *config.AlgorithmConfig) *ELOAlgorithm {
	return &ELOAlgorithm{
		config: config,
		stats: &AlgorithmStats{
			QualityDistribution: make(map[string]int64),
			LastUpdated:         time.Now(),
		},
	}
}

func (e *ELOAlgorithm) Name() string {
	return e.config.Name
}

func (e *ELOAlgorithm) Version() string {
	return e.config.Version
}

func (e *ELOAlgorithm) Description() string {
	return e.config.Description
}

func (e *ELOAlgorithm) SetConfig(config *config.AlgorithmConfig) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.config = config
	return nil
}

func (e *ELOAlgorithm) GetConfig() *config.AlgorithmConfig {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.config
}

func (e *ELOAlgorithm) CalculateMatchScore(ctx context.Context, p1, p2 *Player) (float64, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if err := e.ValidatePlayer(p1); err != nil {
		return 0, fmt.Errorf("player 1 validation failed: %w", err)
	}
	if err := e.ValidatePlayer(p2); err != nil {
		return 0, fmt.Errorf("player 2 validation failed: %w", err)
	}

	// 基础因子计算
	levelScore := e.calculateLevelScore(p1, p2)
	winRateScore := e.calculateWinRateScore(p1, p2)
	pingScore := e.calculatePingScore(p1, p2)
	mmrScore := e.calculateMMRScore(p1, p2)

	// 加权计算总分
	totalScore := e.config.Weights["level"]*levelScore + e.config.Weights["winrate"]*winRateScore + e.config.Weights["ping"]*pingScore + e.config.Weights["mmr"]*mmrScore

	// 基于排队时间动态调整
	if e.config.EnableDynamicAdjustment {
		queueTimeBonus := e.calculateQueueTimeBonus(p1, p2)
		totalScore *= queueTimeBonus
	}
	// 返回最终得分，确定分数在0-1之间
	return math.Max(0, math.Min(1, totalScore)), nil

}

func (e *ELOAlgorithm) FindOptimalMatch(ctx context.Context, player *Player, candidates []*Player) (*MatchResult, error) {
	startTime := time.Now()
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no candidates found")
	}

	type candidateScore struct {
		player *Player
		score  float64
	}

	var scores []candidateScore

	// 计算所有候选者的匹配分数
	for _, candidate := range candidates {

		if candidate.ID == player.ID {
			// 跳过自己
			continue
		}

		score, err := e.CalculateMatchScore(ctx, player, candidate)
		if err != nil {
			continue //跳过计算错误

		}
		// 只考虑超过最低质量阈值的匹配

		if score >= e.config.Thresholds["min_quality"] {
			scores = append(scores, candidateScore{player: candidate, score: score})
		}
	}

	if len(scores) == 0 {
		return nil, fmt.Errorf("no suitable matches found")
	}

	// 找到最高分的匹配
	bestMatch := scores[0]
	for _, candidate := range scores[1:] {
		if candidate.score > bestMatch.score {
			bestMatch = candidate
		}
	}

	// 更新统计信息
	e.updateStats(bestMatch.score)

	return &MatchResult{
		MatchID:    fmt.Sprintf("match_%d_%d_%d", player.ID, bestMatch.player.ID, time.Now().Unix()),
		Players:    []*Player{player, bestMatch.player},
		Quality:    bestMatch.score,
		Confidence: e.calculateConfidence(player, bestMatch.player),
		Algorithm:  e.Name(),
		Metadata: map[string]interface{}{
			"calculation_time": time.Since(startTime),
			"candidates_count": len(candidates),
			"qualified_count":  len(scores),
		},
	}, nil

}

// CalculateMMR 计算新的MMR评级
func (e *ELOAlgorithm) CalculateMMR(ctx context.Context, player *Player, gameResult *GameResult) (float64, error) {
	kFactor := e.config.Parameters["k_factor"].(float64)

	// 计算期望胜率
	expectedScore := 1.0 / (1.0 + math.Pow(10, (gameResult.OpponentMMR-player.MMR)/400))

	// 实际得分
	actualScore := 0.0
	if gameResult.IsWin {
		actualScore = 1.0
	}

	// 考虑表现评分的调整
	performanceAdjustment := (gameResult.Performance - 0.5) * 0.2 // -0.1 到 +0.1 的调整

	// 计算新MMR
	newMMR := player.MMR + kFactor*(actualScore-expectedScore) + performanceAdjustment*kFactor

	// 应用边界限制
	floor := e.config.Parameters["rating_floor"].(float64)
	ceiling := e.config.Parameters["rating_ceiling"].(float64)
	newMMR = math.Max(floor, math.Min(ceiling, newMMR))

	return newMMR, nil
}

func (e *ELOAlgorithm) ValidatePlayer(player *Player) error {
	if player == nil {
		return fmt.Errorf("player cannot be nil")
	}
	if player.ID <= 0 {
		return fmt.Errorf("invalid player ID: %d", player.ID)
	}
	if player.Level < 1 || player.Level > 100 {
		return fmt.Errorf("invalid player level: %d", player.Level)
	}
	if player.WinRate < 0 || player.WinRate > 1 {
		return fmt.Errorf("invalid win rate: %f", player.WinRate)
	}
	if player.Ping < 0 {
		return fmt.Errorf("invalid ping: %d", player.Ping)
	}

	return nil
}

func (e *ELOAlgorithm) calculateLevelScore(p1, p2 *Player) float64 {
	levelDiff := math.Abs(float64(p1.Level - p2.Level))
	maxDiff := float64(e.config.MaxLevelDiff)
	if levelDiff > maxDiff {
		return 0
	}
	return 1.0 - (levelDiff / maxDiff)
}

func (e *ELOAlgorithm) calculateWinRateScore(p1, p2 *Player) float64 {
	winRateDiff := math.Abs(p1.WinRate - p2.WinRate)
	maxDiff := e.config.MaxWinRateDiff
	if winRateDiff > maxDiff {
		return 0
	}
	return 1.0 - (winRateDiff / maxDiff)
}

func (e *ELOAlgorithm) calculatePingScore(p1, p2 *Player) float64 {
	avgPing := float64(p1.Ping+p2.Ping) / 2
	maxPing := float64(e.config.MaxPingDiff)
	if avgPing > maxPing {
		return 0
	}
	return 1.0 - (avgPing / maxPing)
}

func (e *ELOAlgorithm) calculateMMRScore(p1, p2 *Player) float64 {
	mmrDiff := math.Abs(p1.MMR - p2.MMR)
	// 使用正态分布计算MMR匹配度
	return math.Exp(-mmrDiff * mmrDiff / (2 * 200 * 200))
}

func (e *ELOAlgorithm) calculateQueueTimeBonus(p1, p2 *Player) float64 {
	avgQueueTime := (time.Since(p1.QueueTime) + time.Since(p2.QueueTime)) / 2
	maxQueueTime := time.Duration(e.config.MaxQueueTime) * time.Second
	if avgQueueTime <= maxQueueTime {
		return 1.0
	}
	// 超时后逐渐放宽标准
	bonus := 1.0 + (avgQueueTime.Seconds()-maxQueueTime.Seconds())/maxQueueTime.Seconds()*e.config.QueueTimeMultiplier
	return math.Min(3.0, bonus) // 最多放宽3倍
}

func (e *ELOAlgorithm) calculateConfidence(p1, p2 *Player) float64 {
	// 基于游戏数量和近期表现计算置信度
	p1Games := float64(p1.WinCount + p1.LoseCount)
	p2Games := float64(p2.WinCount + p2.LoseCount)

	minGames := math.Min(p1Games, p2Games)
	confidence := math.Min(1.0, minGames/50.0)
	return confidence
}

func (e *ELOAlgorithm) updateStats(quality float64) {

	e.mu.Lock()
	defer e.mu.Unlock()

	e.stats.TotalMatches++
	e.stats.SuccessfulMatches++

	//更新平均质量
	e.stats.AverageMatchQuality = (e.stats.AverageMatchQuality*float64(e.stats.TotalMatches-1) + quality) / float64(e.stats.TotalMatches)
	// 更新质量分布
	qualityBucket := fmt.Sprintf("%.1f-%.1f", math.Floor(quality*10)/10, math.Ceil(quality*10)/10)
	e.stats.QualityDistribution[qualityBucket]++
	e.stats.LastUpdated = time.Now()
}

func (e *ELOAlgorithm) GetStats() *AlgorithmStats {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.stats
}

func (e *ELOAlgorithm) ResetStats() {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.stats = &AlgorithmStats{
		QualityDistribution: make(map[string]int64),
		LastUpdated:         time.Now(),
	}
}
