package api

import (
	"fmt"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/suoten/jt-simulate/internal/api/handler"
	"github.com/suoten/jt-simulate/internal/api/websocket"
	"github.com/suoten/jt-simulate/internal/engine"
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
	// 将WebSocket Hub注入引擎，实现实时消息监控
	e.SetBroadcast(func(direction, phone, msgName, msgID string, raw []byte) {
		hub.BroadcastMessage(direction, phone, msgName, msgID, raw)
	})
	return &Server{
		engine:   e,
		workshop: w,
		hub:      hub,
	}
}

// Start 启动服务器
func (s *Server) Start(host string, port int, mode string, frontendFS fs.FS) error {
	gin.SetMode(mode)
	r := gin.Default()

	// CORS
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// API路由
	h := handler.New(s.engine, s.workshop)
	h.RegisterRoutes(r)

	// WebSocket — 路径必须与前端代码匹配
	r.GET("/api/v1/ws/monitor", func(c *gin.Context) {
		s.hub.HandleWS(c.Writer, c.Request)
	})
	// 兼容旧路径
	r.GET("/ws", func(c *gin.Context) {
		s.hub.HandleWS(c.Writer, c.Request)
	})

	// 前端静态文件
	if frontendFS != nil {
		r.NoRoute(func(c *gin.Context) {
			// 尝试读取静态文件
			path := c.Request.URL.Path
			if path == "/" {
				path = "/index.html"
			}
			data, err := fs.ReadFile(frontendFS, path[1:])
			if err != nil {
				// SPA fallback
				data, err = fs.ReadFile(frontendFS, "index.html")
				if err != nil {
					c.String(http.StatusNotFound, "Not Found")
					return
				}
				c.Data(http.StatusOK, "text/html; charset=utf-8", data)
				return
			}
			// 根据扩展名设置Content-Type
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
	return r.Run(addr)
}

// Hub 获取WebSocket Hub
func (s *Server) Hub() *websocket.Hub {
	return s.hub
}
