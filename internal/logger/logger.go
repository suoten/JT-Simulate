package logger

import (
	"log/slog"
	"os"
	"sync/atomic"
)

// defaultLogger 全局 logger（原子读写：SetLogger 与 L 可能并发调用）
var defaultLogger atomic.Pointer[slog.Logger]

func init() {
	defaultLogger.Store(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))
}

// L 返回全局 logger
func L() *slog.Logger {
	return defaultLogger.Load()
}

// SetLogger 设置全局 logger
func SetLogger(l *slog.Logger) {
	if l == nil {
		return
	}
	defaultLogger.Store(l)
}

// Info Info级别日志
func Info(msg string, args ...any) { L().Info(msg, args...) }

// Warn Warn级别日志
func Warn(msg string, args ...any) { L().Warn(msg, args...) }

// Error Error级别日志
func Error(msg string, args ...any) { L().Error(msg, args...) }

// Debug Debug级别日志
func Debug(msg string, args ...any) { L().Debug(msg, args...) }
