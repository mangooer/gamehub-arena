package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Database   DatabaseConfig   `mapstructure:"database"`
	Redis      RedisConfig      `mapstructure:"redis"`
	Logging    LoggingConfig    `mapstructure:"logging"`
	Auth       AuthConfig       `mapstructure:"auth"`
	Monitoring MonitoringConfig `mapstructure:"monitoring"`
	Match      MatchConfig      `mapstructure:"match"`
}

type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type DatabaseConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Name         string `mapstructure:"name"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	SSLMode      string `mapstructure:"ssl_mode"`       // 新增
	MaxOpenConns int    `mapstructure:"max_open_conns"` // 新增
	MaxIdleConns int    `mapstructure:"max_idle_conns"` // 新增
	MaxLifetime  int    `mapstructure:"max_lifetime"`   // 新增（小时）
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type LoggingConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`      // 输出方式：console, file, both
	FilePath   string `mapstructure:"file_path"`   // 日志文件路径
	MaxSize    int    `mapstructure:"max_size"`    // 日志文件最大大小(MB)
	MaxBackups int    `mapstructure:"max_backups"` // 保留的日志文件数量
	MaxAge     int    `mapstructure:"max_age"`     // 日志文件保留天数
	Compress   bool   `mapstructure:"compress"`    // 是否压缩旧日志文件
}

type AuthConfig struct {
	Jwt      JwtConfig      `mapstructure:"jwt"`
	Password PasswordConfig `mapstructure:"password"`
	Session  SessionConfig  `mapstructure:"session"`
}

type JwtConfig struct {
	SecretKey       string `mapstructure:"secret_key"`
	AccessTokenTTL  int    `mapstructure:"access_token_ttl"`
	RefreshTokenTTL int    `mapstructure:"refresh_token_ttl"`
	Issuer          string `mapstructure:"issuer"`
	Audience        string `mapstructure:"audience"`
}

type PasswordConfig struct {
	MinLength     int  `mapstructure:"min_length"`
	RequireUpper  bool `mapstructure:"require_upper"`
	RequireLower  bool `mapstructure:"require_lower"`
	RequireDigit  bool `mapstructure:"require_digit"`
	RequireSymbol bool `mapstructure:"require_symbol"`
	BcryptCost    int  `mapstructure:"bcrypt_cost"`
}

type SessionConfig struct {
	MaxSessions     int `mapstructure:"max_sessions"`     // 每用户最大会话数
	InactiveTimeout int `mapstructure:"inactive_timeout"` // 非活跃超时（分钟）
	RememberMeDays  int `mapstructure:"remember_me_days"` // 记住我天数
}

// 监控配置
type MonitoringConfig struct {
	Metrics MetricsConfig `mapstructure:"metrics"`
	Tracing TracingConfig `mapstructure:"tracing"`
	Health  HealthConfig  `mapstructure:"health"`
}

type MetricsConfig struct {
	Enabled   bool   `mapstructure:"enabled"`
	Port      int    `mapstructure:"port"`
	Path      string `mapstructure:"path"`
	Namespace string `mapstructure:"namespace"`
	Subsystem string `mapstructure:"subsystem"`
}

type TracingConfig struct {
	Enabled     bool    `mapstructure:"enabled"`
	ServiceName string  `mapstructure:"service_name"`
	Endpoint    string  `mapstructure:"endpoint"`
	SampleRate  float64 `mapstructure:"sample_rate"`
}

type HealthConfig struct {
	Enabled       bool   `mapstructure:"enabled"`
	Port          int    `mapstructure:"port"`
	Path          string `mapstructure:"path"`
	CheckInterval int    `mapstructure:"check_interval"` // 健康检查间隔(秒)
}

type MatchConfig struct {
	DefaultAlgorithm string                     `mapstructure:"default_algorithm"`
	Algorithms       map[string]AlgorithmConfig `mapstructure:"algorithms"`
}

type AlgorithmConfig struct {
	Name        string                 `mapstructure:"name"`
	Description string                 `mapstructure:"description"`
	Version     string                 `mapstructure:"version"`
	Enabled     bool                   `mapstructure:"enabled"`
	Weights     map[string]float64     `mapstructure:"weights"`    // 各因子权重
	Thresholds  map[string]float64     `mapstructure:"thresholds"` //阈值配置
	Parameters  map[string]interface{} `mapstructure:"parameters"` //算法特定参数配置

	//匹配限制
	MaxLevelDiff   int     `mapstructure:"max_level_diff"`    //最大等级差
	MaxWinRateDiff float64 `mapstructure:"max_win_rate_diff"` //最大胜率差
	MaxPingDiff    int     `mapstructure:"max_ping_diff"`     //最大延迟差
	MaxQueueTime   int     `mapstructure:"max_queue_time"`    //最大排队时间 秒

	//动态调整
	EnableDynamicAdjustment bool    `mapstructure:"enable_dynamic_adjustment"` //是否启用动态调整
	QueueTimeMultiplier     float64 `mapstructure:"queue_time_multiplier"`     //排队时间倍率
}

func Load() (*Config, error) {

	// 设置默认值
	viper.SetDefault("database.ssl_mode", "disable")
	viper.SetDefault("database.max_open_conns", 100)
	viper.SetDefault("database.max_idle_conns", 10)
	viper.SetDefault("database.max_lifetime", 10)

	// 认证相关默认值
	viper.SetDefault("auth.jwt.access_token_ttl", 15)   // 15分钟
	viper.SetDefault("auth.jwt.refresh_token_ttl", 168) // 7天（小时）
	viper.SetDefault("auth.jwt.issuer", "gamehub-arena")
	viper.SetDefault("auth.jwt.audience", "gamehub-users")

	viper.SetDefault("auth.password.min_length", 8)
	viper.SetDefault("auth.password.require_upper", true)
	viper.SetDefault("auth.password.require_lower", true)
	viper.SetDefault("auth.password.require_digit", true)
	viper.SetDefault("auth.password.require_symbol", false)
	viper.SetDefault("auth.password.bcrypt_cost", 12)

	viper.SetDefault("auth.session.max_sessions", 5)
	viper.SetDefault("auth.session.inactive_timeout", 30) // 30分钟
	viper.SetDefault("auth.session.remember_me_days", 30)

	// 日志相关默认值
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")
	viper.SetDefault("logging.output", "console")
	viper.SetDefault("logging.file_path", "./logs/app.log")
	viper.SetDefault("logging.max_size", 100) // 100MB
	viper.SetDefault("logging.max_backups", 3)
	viper.SetDefault("logging.max_age", 28) // 28天
	viper.SetDefault("logging.compress", true)

	// 监控相关默认值
	viper.SetDefault("monitoring.metrics.enabled", true)
	viper.SetDefault("monitoring.metrics.port", 9090)
	viper.SetDefault("monitoring.metrics.path", "/metrics")
	viper.SetDefault("monitoring.metrics.namespace", "gamehub")
	viper.SetDefault("monitoring.metrics.subsystem", "arena")

	viper.SetDefault("monitoring.tracing.enabled", false)
	viper.SetDefault("monitoring.tracing.service_name", "gamehub-arena")
	viper.SetDefault("monitoring.tracing.endpoint", "http://localhost:14268/api/traces")
	viper.SetDefault("monitoring.tracing.sample_rate", 0.1)

	viper.SetDefault("monitoring.health.enabled", true)
	viper.SetDefault("monitoring.health.port", 8080)
	viper.SetDefault("monitoring.health.path", "/health")
	viper.SetDefault("monitoring.health.check_interval", 30)

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
