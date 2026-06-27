package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Logger 封装 slog.Logger
type Logger struct {
	*slog.Logger
	level slog.Level
}

// New 创建日志实例
func New(cfg Config) (*Logger, error) {
	level, err := parseLevel(cfg.Level)
	if err != nil {
		return nil, err
	}

	// 确保日志目录存在
	if cfg.OutputPath != "" && cfg.OutputPath != "stdout" {
		dir := filepath.Dir(cfg.OutputPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("create log directory: %w", err)
		}
	}

	var output io.Writer = os.Stdout
	if cfg.OutputPath != "" && cfg.OutputPath != "stdout" {
		file, err := os.OpenFile(cfg.OutputPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("open log file: %w", err)
		}
		output = io.MultiWriter(os.Stdout, file)
	}

	replaceAttr := func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.TimeKey {
			if t, ok := a.Value.Any().(time.Time); ok {
				a.Value = slog.StringValue(t.Format("2006-01-02 15:04:05.0000000"))
			}
		}
		return a
	}

	var handler slog.Handler
	if strings.ToLower(cfg.Format) == "json" {
		handler = slog.NewJSONHandler(output, &slog.HandlerOptions{
			Level:       level,
			ReplaceAttr: replaceAttr,
		})
	} else {
		handler = slog.NewTextHandler(output, &slog.HandlerOptions{
			Level:       level,
			ReplaceAttr: replaceAttr,
		})
	}

	return &Logger{
		Logger: slog.New(handler),
		level:  level,
	}, nil
}

// parseLevel 解析日志级别
func parseLevel(level string) (slog.Level, error) {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("unknown log level: %s", level)
	}
}

// Config 日志配置
type Config struct {
	Level      string
	Format     string
	OutputPath string
}

// WithComponent 添加组件字段
func (l *Logger) WithComponent(component string) *Logger {
	return &Logger{
		Logger: l.Logger.With("component", component),
		level:  l.level,
	}
}

// WithError 添加错误字段
func (l *Logger) WithError(err error) *Logger {
	return &Logger{
		Logger: l.Logger.With("error", err.Error()),
		level:  l.level,
	}
}

// Fatal 记录致命错误并退出
func (l *Logger) Fatal(msg string, args ...any) {
	l.Error(msg, args...)
	os.Exit(1)
}

// TimeFormat 返回标准时间格式字符串
func TimeFormat(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}
