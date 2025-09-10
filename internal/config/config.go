package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Logging  LoggingConfig  `mapstructure:"logging"`
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
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
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
