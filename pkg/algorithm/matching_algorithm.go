package algorithm

import (
	"context"
	"time"

	"github.com/mangooer/gamehub-arena/internal/config"
)

type Player struct {
	ID        uint64    `json:"id"`
	Username  string    `json:"username"`
	Level     int       `json:"level"`
	Rank      string    `json:"rank"`
	WinRate   float64   `json:"win_rate"`
	WinCount  int       `json:"win_count"`
	LoseCount int       `json:"lose_count"`
	Ping      int       `json:"ping"`
	QueueTime time.Time `json:"queue_time"`
	GameMode  string    `json:"game_mode"`
	Region    string    `json:"region"`
	// 扩展字段，用于算法计算
	MMR         float64      `json:"mmr"`          //匹配评级
	Confidence  float64      `json:"confidence"`   //评级置信度
	RecentGames []GameResult `json:"recent_games"` //最近游戏结果
}

// 近期游戏结果
type GameResult struct {
	IsWin       bool      `json:"is_win"`
	GameTime    time.Time `json:"game_time"`
	OpponentMMR float64   `json:"opponent_mmr"`
	Performance float64   `json:"performance"`
}

// 匹配结果
type MatchResult struct {
	MatchID    string                 `json:"match_id"`
	Players    []*Player              `json:"players"`
	Quality    float64                `json:"quality"`    //匹配质量
	Confidence float64                `json:"confidence"` //匹配置信度
	Algorithm  string                 `json:"algorithm"`  //匹配算法
	Metadata   map[string]interface{} `json:"metadata"`   //匹配元数据
}

// 算法统计信息
type AlgorithmStats struct {
	TotalMatches        int64         `json:"total_matches"`         //匹配总数
	SuccessfulMatches   int64         `json:"successful_matches"`    //成功匹配数
	AverageMatchTime    time.Duration `json:"average_match_time"`    //平均匹配时间
	AverageMatchQuality float64       `json:"average_match_quality"` //平均匹配质量
	LastUpdated         time.Time     `json:"last_updated"`          //最后更新时间

	// 性能指标
	CalculationsPerSecond float64 `json:"calculations_per_second"` //每秒计算次数
	ErrorRate             float64 `json:"error_rate"`              //错误率

	// 匹配质量分布
	QualityDistribution map[string]int64 `json:"quality_distribution"` //匹配质量分布

}

// 匹配算法接口
type MatchingAlgorithm interface {
	// 算法基本信息
	Name() string
	Version() string
	Description() string

	// 核心匹配功能
	CalculateMatchScore(ctx context.Context, p1, p2 *Player) (float64, error)
	FindOptimalMatch(ctx context.Context, player *Player, candidates []*Player) (*MatchResult, error)
	CalculateMMR(ctx context.Context, player *Player, gameResult *GameResult) (float64, error)

	// 配置和调优
	SetConfig(config *config.AlgorithmConfig) error
	GetConfig() *config.AlgorithmConfig
	ValidatePlayer(Player *Player) error

	//性能监控
	GetStats() *AlgorithmStats
	ResetStats()
}
