package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"net"
	"os"

	"github.com/spf13/cobra"
	"github.com/suoten/jt-simulate/internal/api"
	"github.com/suoten/jt-simulate/internal/config"
	"github.com/suoten/jt-simulate/internal/engine"
	"github.com/suoten/jt-simulate/internal/logger"
	"github.com/suoten/jt-simulate/internal/workshop"
)

//go:embed all:cmd/jt-simulate/frontend/dist
var frontendDist embed.FS

var (
	configPath string
	version    = "1.0.0"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "jt-simulate",
		Short: "JT-Simulate — 部标协议通用仿真平台",
		Long:  "JT-Simulate 是做部标协议开发、车联网开发的人都该用的仿真工具。",
	}

	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "配置文件路径")

	// serve 子命令 — Web/服务模式
	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "以 Web 服务模式运行（监听端口，浏览器访问）",
		Run:   runServer,
	}
	rootCmd.AddCommand(serveCmd)

	// analyze 子命令 — 分析报文
	analyzeCmd := &cobra.Command{
		Use:   "analyze [hex]",
		Short: "分析报文",
		Args:  cobra.ExactArgs(1),
		Run:   runAnalyze,
	}
	rootCmd.AddCommand(analyzeCmd)

	// version 子命令
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "显示版本信息",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("JT-Simulate v%s\n", version)
		},
	}
	rootCmd.AddCommand(versionCmd)

	// 默认行为：如果没有子命令，启动桌面模式（Wails 原生窗口）
	rootCmd.Run = runDesktop

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// getFrontendFS 获取前端文件系统
func getFrontendFS() fs.FS {
	distFS, err := fs.Sub(frontendDist, "cmd/jt-simulate/frontend/dist")
	if err != nil {
		return nil
	}
	return distFS
}

// setupLogger 根据配置初始化日志
func setupLogger(cfg *config.LogConfig) {
	level := slog.LevelInfo
	switch cfg.Level {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	var handler slog.Handler
	opts := &slog.HandlerOptions{Level: level}
	if cfg.Format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}
	logger.SetLogger(slog.New(handler))
}

// runDesktop 启动 Wails 桌面应用（双击 exe 默认行为）
func runDesktop(cmd *cobra.Command, args []string) {
	// 桌面模式使用默认配置
	cfg := config.DefaultConfig()
	setupLogger(&cfg.Log)

	eng := engine.New()
	if err := eng.LoadDevices(); err != nil {
		logger.Error("加载持久化设备失败", "error", err)
	}
	ws := workshop.New()
	server := api.NewServer(eng, ws)
	frontendFS := getFrontendFS()

	// 找一个空闲端口启动后台 HTTP 服务
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "无法分配端口: %v\n", err)
		os.Exit(1)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	// 后台启动 HTTP 服务（不提供前端文件，由 Wails 提供）
	go func() {
		if err := server.Start("127.0.0.1", port, "release", nil); err != nil {
			logger.Error("桌面服务启动失败", "error", err)
			os.Exit(1)
		}
	}()

	serverURL := fmt.Sprintf("http://127.0.0.1:%d", port)

	// 启动 Wails 桌面窗口
	if err := runWailsDesktop(serverURL, frontendFS); err != nil {
		fmt.Fprintf(os.Stderr, "桌面应用启动失败: %v\n", err)
		os.Exit(1)
	}
}

// runServer 运行 Web 服务
func runServer(cmd *cobra.Command, args []string) {
	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 根据配置初始化日志
	setupLogger(&cfg.Log)

	logger.Info("JT-Simulate 启动", "version", version)
	logger.Info("服务模式", "host", cfg.Server.Host, "port", cfg.Server.Port)

	eng := engine.New()
	if err := eng.LoadDevices(); err != nil {
		logger.Error("加载持久化设备失败", "error", err)
	}
	ws := workshop.New()

	server := api.NewServer(eng, ws)

	frontendFS := getFrontendFS()
	if frontendFS == nil {
		logger.Warn("前端文件未找到，仅提供API服务")
	}

	logger.Info("按 Ctrl+C 优雅退出")

	if cfg.Server.CertFile != "" && cfg.Server.KeyFile != "" {
		logger.Info("启用 HTTPS", "cert", cfg.Server.CertFile)
		if err := server.StartWithTLS(cfg.Server.Host, cfg.Server.Port, cfg.Server.Mode, frontendFS, cfg.Server.CertFile, cfg.Server.KeyFile); err != nil {
			logger.Error("服务启动失败", "error", err)
			os.Exit(1)
		}
	} else {
		if err := server.Start(cfg.Server.Host, cfg.Server.Port, cfg.Server.Mode, frontendFS); err != nil {
			logger.Error("服务启动失败", "error", err)
			os.Exit(1)
		}
	}
}

// runAnalyze 运行分析
func runAnalyze(cmd *cobra.Command, args []string) {
	ws := workshop.New()
	resp := ws.Analyze(&workshop.AnalyzeRequest{
		Protocol: "jt808",
		Hex:      args[0],
	})

	if !resp.Success {
		fmt.Printf("分析失败: %s\n", resp.Error)
		os.Exit(1)
	}

	fmt.Printf("消息ID: %s (%s)\n", resp.MsgID, resp.MsgName)
	fmt.Printf("终端手机号: %s\n", resp.Phone)
	fmt.Printf("流水号: %d\n", resp.SeqNum)
	fmt.Printf("解析结果: %+v\n", resp.Parsed)
}
