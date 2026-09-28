package api

import (
	"context"
	"crypto/subtle"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/suoten/jt-simulate/internal/api/handler"
	"github.com/suoten/jt-simulate/internal/api/websocket"
	"github.com/suoten/jt-simulate/internal/engine"
	"github.com/suoten/jt-simulate/internal/logger"
	"github.com/suoten/jt-simulate/internal/workshop"
)

// Server API服务器
type Server struct {
	engine   *engine.Engine
	workshop *workshop.Workshop
	hub      *websocket.Hub
}

// NewServer 创建服务器
func NewServer(e *engine.Engine, w *workshop.Workshop) *Server {
	hub := websocket.NewHub()
	e.SetBroadcast(func(direction, phone, msgName, msgID string, raw []byte) {
		hub.BroadcastMessage(direction, phone, msgName, msgID, raw)
	})
	return &Server{
		engine:   e,
		workshop: w,
		hub:      hub,
	}
}

// Start 启动服务器（支持优雅关闭）
func (s *Server) Start(host string, port int, mode string, frontendFS fs.FS) error {
	gin.SetMode(mode)
	r := gin.Default()

	// CORS — 桌面模式（127.0.0.1）不需要认证；服务模式需要认证
	isDesktop := host == "127.0.0.1" || host == "localhost"

	r.Use(func(c *gin.Context) {
		if isDesktop {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		} else {
			origin := c.Request.Header.Get("Origin")
			if origin != "" {
				c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			}
		}
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// API 认证中间件（桌面模式跳过）
	if !isDesktop {
		apiToken := os.Getenv("JT_SIMULATE_TOKEN")
		if apiToken == "" {
			apiToken = "jt-simulate-default-token"
			logger.Warn("未设置 JT_SIMULATE_TOKEN 环境变量，使用默认 token，生产环境请务必设置",
				"default_token", apiToken)
		}
		r.Use(authMiddleware(apiToken))
	}

	// 限流中间件（每秒最多 20 个请求，突发 50）
	rl := newRateLimiter(20, 50)
	go rl.cleanupLoop() // 后台清理过期 bucket
	r.Use(rl.middleware())

	// API路由
	h := handler.New(s.engine, s.workshop)
	h.RegisterRoutes(r)

	// WebSocket
	r.GET("/api/v1/ws/monitor", func(c *gin.Context) {
		s.hub.HandleWS(c.Writer, c.Request)
	})
	r.GET("/ws", func(c *gin.Context) {
		s.hub.HandleWS(c.Writer, c.Request)
	})

	// 前端静态文件
	if frontendFS != nil {
		r.NoRoute(func(c *gin.Context) {
			path := c.Request.URL.Path
			if path == "/" {
				path = "/index.html"
			}
			data, err := fs.ReadFile(frontendFS, path[1:])
			if err != nil {
				data, err = fs.ReadFile(frontendFS, "index.html")
				if err != nil {
					c.String(http.StatusNotFound, "Not Found")
					return
				}
				c.Data(http.StatusOK, "text/html; charset=utf-8", data)
				return
			}
			contentType := "application/octet-stream"
			switch {
			case strings.HasSuffix(path, ".html"):
				contentType = "text/html; charset=utf-8"
			case strings.HasSuffix(path, ".js"):
				contentType = "application/javascript"
			case strings.HasSuffix(path, ".css"):
				contentType = "text/css"
			case strings.HasSuffix(path, ".json"):
				contentType = "application/json"
			case strings.HasSuffix(path, ".svg"):
				contentType = "image/svg+xml"
			case strings.HasSuffix(path, ".png"):
				contentType = "image/png"
			case strings.HasSuffix(path, ".ico"):
				contentType = "image/x-icon"
			}
			c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
			c.Data(http.StatusOK, contentType, data)
		})
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	// 启动 HTTP 服务（异步）
	go func() {
		logger.Info("HTTP 服务启动", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP 服务异常", "error", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("正在关闭服务...")

	// 停止所有设备
	s.engine.StopAll()

	// 优雅关闭 HTTP 服务（5秒超时）
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("HTTP 服务关闭异常", "error", err)
		return err
	}

	logger.Info("服务已关闭")
	return nil
}

// Hub 获取WebSocket Hub
func (s *Server) Hub() *websocket.Hub {
	return s.hub
}

// authMiddleware API token 认证中间件（使用 constant-time comparison 防时序攻击）
func authMiddleware(token string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 健康检查和 WebSocket 不需要认证
		path := c.Request.URL.Path
		if path == "/api/v1/health" || strings.HasPrefix(path, "/ws") || strings.HasPrefix(path, "/api/v1/ws") {
			c.Next()
			return
		}

		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "缺少认证信息"})
			c.Abort()
			return
		}

		// 支持 "Bearer <token>" 和直接 token 两种格式
		tokenVal := auth
		if strings.HasPrefix(auth, "Bearer ") {
			tokenVal = strings.TrimPrefix(auth, "Bearer ")
		}

		// constant-time comparison 防止时序攻击
		if subtle.ConstantTimeCompare([]byte(tokenVal), []byte(token)) != 1 {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "认证失败"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// rateLimiter 令牌桶限流器（带过期清理）
type rateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*bucket
	rate    float64
	burst   float64
}

type bucket struct {
	tokens float64
	last   time.Time
}

func newRateLimiter(rate, burst int) *rateLimiter {
	return &rateLimiter{
		buckets: make(map[string]*bucket),
		rate:    float64(rate),
		burst:   float64(burst),
	}
}

func (rl *rateLimiter) middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		rl.mu.Lock()
		b, ok := rl.buckets[ip]
		if !ok {
			b = &bucket{tokens: rl.burst, last: time.Now()}
			rl.buckets[ip] = b
		}
		elapsed := time.Since(b.last).Seconds()
		b.tokens += elapsed * rl.rate
		if b.tokens > rl.burst {
			b.tokens = rl.burst
		}
		b.last = time.Now()
		if b.tokens < 1 {
			rl.mu.Unlock()
			c.JSON(http.StatusTooManyRequests, gin.H{"success": false, "error": "请求过于频繁"})
			c.Abort()
			return
		}
		b.tokens--
		rl.mu.Unlock()
		c.Next()
	}
}

// cleanupLoop 定期清理过期的 bucket（每 5 分钟，超过 10 分钟未访问的 IP 清除）
func (rl *rateLimiter) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		cutoff := time.Now().Add(-10 * time.Minute)
		for ip, b := range rl.buckets {
			if b.last.Before(cutoff) {
				delete(rl.buckets, ip)
			}
		}
		rl.mu.Unlock()
	}
}
