package logger

import (
	"os"
	"path/filepath"

	"github.com/mangooer/gamehub-arena/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Logger struct {
	logger *zap.Logger
}

func NewLogger(cfg *config.LoggingConfig) (*Logger, error) {

	var core zapcore.Core

	// 配置编码器
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	var encoder zapcore.Encoder

	if cfg.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}
	// 配置日志级别

	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		level = zapcore.InfoLevel
	}

	// 配置输出
	var writeSyncer zapcore.WriteSyncer
	switch cfg.Output {
	case "file":
		writeSyncer = getFileWriter(cfg)
	case "both":
		writeSyncer = zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), getFileWriter(cfg))
	default:
		writeSyncer = zapcore.AddSync(os.Stdout)
	}

	core = zapcore.NewCore(encoder, writeSyncer, level)

	// 添加调用者信息
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	return &Logger{logger: logger}, nil
}

func getFileWriter(cfg *config.LoggingConfig) zapcore.WriteSyncer {
	// 确保文件目录存在
	if err := os.MkdirAll(filepath.Dir(cfg.FilePath), 0755); err != nil {
		panic(err)
	}
	// 配置日志轮转
	lumberjackLogger := &lumberjack.Logger{
		Filename:   cfg.FilePath,
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge,
		Compress:   cfg.Compress,
	}
	return zapcore.AddSync(lumberjackLogger)
}

// 创建子日志记录器
func (l *Logger) With(fields ...zap.Field) *Logger {
	return &Logger{logger: l.logger.With(fields...)}
}

// 创建命名日志器
func (l *Logger) Named(name string) *Logger {
	return &Logger{logger: l.logger.Named(name)}
}

// 关闭日志器
func (l *Logger) Close() error {
	return l.logger.Sync()
}

func (l *Logger) GetLogger() *zap.Logger {
	return l.logger
}
