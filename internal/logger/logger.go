package logger

import (
	"log/slog"
	"os"
	"sync"
)

var (
	defaultLogger *slog.Logger
	once          sync.Once
)

// L 返回全局 logger
func L() *slog.Logger {
	once.Do(func() {
		defaultLogger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
	})
	return defaultLogger
}

// SetLogger 设置全局 logger
func SetLogger(l *slog.Logger) {
	defaultLogger = l
}

// Info Info级别日志
func Info(msg string, args ...any) { L().Info(msg, args...) }

// Warn Warn级别日志
func Warn(msg string, args ...any) { L().Warn(msg, args...) }

// Error Error级别日志
func Error(msg string, args ...any) { L().Error(msg, args...) }

// Debug Debug级别日志
func Debug(msg string, args ...any) { L().Debug(msg, args...) }
