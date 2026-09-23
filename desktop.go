package main

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

// waitForServer 等待 HTTP 服务就绪
func waitForServer(url string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get(url + "/api/v1/stats")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			return true
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(100 * time.Millisecond)
	}
	return false
}

// runWailsDesktop 启动 Wails 桌面窗口
// 前端静态文件由 Wails AssetServer 直接提供
// API 请求 (/api/) 通过反向代理转发到本地 HTTP 服务
func runWailsDesktop(serverURL string, frontendFS fs.FS) error {
	targetURL, _ := url.Parse(serverURL)
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	return wails.Run(&options.App{
		Title:     "JT-Simulate — 部标协议仿真平台",
		Width:     1440,
		Height:    900,
		MinWidth:  1024,
		MinHeight: 700,
		AssetServer: &assetserver.Options{
			Assets: frontendFS,
			Middleware: func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// API 和 WebSocket 请求代理到后端 HTTP 服务
					if len(r.URL.Path) >= 4 && r.URL.Path[:4] == "/api" {
						proxy.ServeHTTP(w, r)
						return
					}
					// 其他请求由 Wails 默认处理器提供静态文件
					next.ServeHTTP(w, r)
				})
			},
		},
		OnStartup: func(ctx context.Context) {
			if !waitForServer(serverURL, 5*time.Second) {
				fmt.Fprintln(os.Stderr, "警告: 后台服务未就绪")
			}
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
		},
	})
}
