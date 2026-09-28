package handler

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/suoten/jt-simulate/internal/checker"
	"github.com/suoten/jt-simulate/internal/engine"
	"github.com/suoten/jt-simulate/internal/logger"
	gbt32960sim "github.com/suoten/jt-simulate/internal/simulator/gbt32960"
	jt808sim "github.com/suoten/jt-simulate/internal/simulator/jt808"
	"github.com/suoten/jt-simulate/internal/simulator/base"
	"github.com/suoten/jt-simulate/internal/workshop"
)

// Handler API处理器
type Handler struct {
	engine         *engine.Engine
	workshop       *workshop.Workshop
	checker        *checker.Checker
	fuzzer         *checker.Fuzzer
	scenarioEngine *engine.ScenarioEngine
	stressEngine   *engine.StressEngine
}

// New 创建处理器
func New(e *engine.Engine, w *workshop.Workshop) *Handler {
	se := engine.NewScenarioEngine()
	for _, s := range engine.DefaultScenarios() {
		se.Register(s)
	}
	return &Handler{
		engine:         e,
		workshop:       w,
		checker:        checker.New(),
		fuzzer:         checker.NewFuzzer(),
		scenarioEngine: se,
		stressEngine:   engine.NewStressEngine(),
	}
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")
	{
		// 报文工坊
		api.POST("/workshop/encode", h.EncodeMessage)
		api.POST("/workshop/analyze", h.AnalyzeMessage)

		// 设备管理
		api.GET("/devices", h.ListDevices)
		api.POST("/devices", h.CreateDevice)
		api.GET("/devices/:id", h.GetDevice)
		api.POST("/devices/:id/start", h.StartDevice)
		api.POST("/devices/:id/stop", h.StopDevice)
		api.DELETE("/devices/:id", h.DeleteDevice)

		// 合规检查
		api.POST("/check/compliance", h.CheckCompliance)
		api.POST("/check/fuzz", h.RunFuzz)

		// 场景管理
		api.GET("/scenarios", h.ListScenarios)
		api.GET("/scenarios/:name", h.GetScenario)
		api.POST("/scenarios/run", h.RunScenario)
		api.POST("/scenarios/stop", h.StopScenario)
		api.GET("/scenarios/status", h.ScenarioStatus)

		// 压测
		api.POST("/stress/run", h.RunStress)
		api.POST("/stress/stop", h.StopStress)
		api.GET("/stress/status", h.StressStatus)

		// 引擎统计
		api.GET("/stats", h.GetStats)

		// 健康检查
		api.GET("/health", h.Health)
	}
}

// ============================== 报文工坊 ==============================

func (h *Handler) EncodeMessage(c *gin.Context) {
	var req workshop.EncodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	resp := h.workshop.Encode(&req)
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) AnalyzeMessage(c *gin.Context) {
	var req workshop.AnalyzeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	resp := h.workshop.Analyze(&req)
	c.JSON(http.StatusOK, resp)
}

// ============================== 设备管理 ==============================

func (h *Handler) ListDevices(c *gin.Context) {
	devices := h.engine.ListDevices()
	list := make([]gin.H, 0, len(devices))
	for _, d := range devices {
		list = append(list, gin.H{
			"id":         d.Config.ID,
			"protocol":   d.Config.Protocol,
			"phone":      d.Config.Phone,
			"plate":      d.Config.Plate,
			"state":      d.State().String(),
			"target":     d.Config.TargetAddr,
			"interval":   d.Config.LocationInterval,
			"heartbeat":  d.Config.HeartbeatInterval,
		})
	}
	c.JSON(http.StatusOK, gin.H{"devices": list})
}

func (h *Handler) CreateDevice(c *gin.Context) {
	var cfg base.DeviceConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	if cfg.ID == "" {
		cfg.ID = cfg.Phone
	}
	if err := h.engine.CreateDevice(&cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "id": cfg.ID})
}

func (h *Handler) GetDevice(c *gin.Context) {
	id := c.Param("id")
	info, err := h.engine.GetDevice(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":       info.Config.ID,
		"protocol": info.Config.Protocol,
		"phone":    info.Config.Phone,
		"plate":    info.Config.Plate,
		"state":    info.State().String(),
		"target":   info.Config.TargetAddr,
		"config":   info.Config,
	})
}

func (h *Handler) StartDevice(c *gin.Context) {
	id := c.Param("id")
	if err := h.engine.StartDevice(c.Request.Context(), id); err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "connectex: No connection") || strings.Contains(errMsg, "connection refused") {
			c.JSON(http.StatusOK, gin.H{"success": false, "error": "无法连接到目标平台，请确认目标地址是否正确且平台已启动", "error_type": "connection_failed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": false, "error": errMsg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "设备已启动"})
}

func (h *Handler) StopDevice(c *gin.Context) {
	id := c.Param("id")
	if err := h.engine.StopDevice(id); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "设备已停止"})
}

func (h *Handler) DeleteDevice(c *gin.Context) {
	id := c.Param("id")
	if err := h.engine.DeleteDevice(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ============================== 合规检查 ==============================

func (h *Handler) CheckCompliance(c *gin.Context) {
	var req struct {
		Protocol string `json:"protocol"`
		Hex      string `json:"hex"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	if req.Protocol == "" {
		req.Protocol = "jt808"
	}

	result, err := h.checker.CheckRaw(req.Protocol, req.Hex)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"result":  result,
	})
}

func (h *Handler) RunFuzz(c *gin.Context) {
	result := h.fuzzer.Run()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"result":  result,
	})
}

// ============================== 场景管理 ==============================

func (h *Handler) ListScenarios(c *gin.Context) {
	scenarios := h.scenarioEngine.List()
	c.JSON(http.StatusOK, gin.H{"scenarios": scenarios})
}

func (h *Handler) GetScenario(c *gin.Context) {
	name := c.Param("name")
	s, err := h.scenarioEngine.Get(name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, s)
}

func (h *Handler) RunScenario(c *gin.Context) {
	var req struct {
		Name   string `json:"name"`
		Target string `json:"target"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	if h.scenarioEngine.IsRunning() {
		c.JSON(http.StatusConflict, gin.H{"success": false, "error": "已有场景正在运行，请先停止"})
		return
	}

	scenario, err := h.scenarioEngine.Get(req.Name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	targetAddr := req.Target
	if targetAddr == "" {
		targetAddr = "127.0.0.1:7611"
	}

	phone := "013800009999"
	devCfg := &base.DeviceConfig{
		ID:                "scenario_" + phone,
		Protocol:          scenario.Protocol,
		Phone:             phone,
		Plate:             "场景测试",
		PlateColor:        1,
		TargetAddr:        targetAddr,
		AuthCode:          "scenario_test",
		ProvinceID:        11,
		CityID:            100,
		Manufacturer:      "SUOTEN",
		TerminalModel:     "JT-100",
		TerminalID:        "9999999",
		HeartbeatInterval: 60,
		LocationInterval:  5,
		ReconnectInterval: 15,
		StartLat:          39.9093,
		StartLon:          116.3974,
	}

	// 创建仿真器
	var sim engine.Simulator
	switch scenario.Protocol {
	case "gbt32960":
		sim = gbt32960sim.New(devCfg)
	default:
		sim = jt808sim.New(devCfg)
	}

	sim.SetOnSend(func(data []byte) {
		// 场景设备发送的消息不计入引擎统计
	})

	ctx := context.Background()
	if err := sim.Online(ctx); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "error": fmt.Sprintf("设备连接失败: %v", err)})
		return
	}

	// 异步执行场景步骤
	go func() {
		defer sim.Offline()

		err := h.scenarioEngine.Run(ctx, req.Name, func(step *engine.ScenarioStep) error {
			logger.Info("场景步骤", "name", step.Name, "action", step.Action)

			switch step.Action {
			case "send":
				msgIDStr, _ := step.Params["msg_id"].(string)
				logger.Debug("场景发送消息", "msg_id", msgIDStr)

			case "location":
				speedF, _ := step.Params["speed"].(float64)
				if speedF == 0 {
					if speedI, ok := step.Params["speed"].(int); ok {
						speedF = float64(speedI)
					}
				}
				lat, _ := step.Params["lat"].(float64)
				lon, _ := step.Params["lon"].(float64)

				// 安全类型断言，避免 panic
				s, ok := sim.(*jt808sim.Simulator)
				if !ok {
					logger.Warn("场景步骤跳过：非JT808仿真器不支持位置操作", "protocol", scenario.Protocol)
					break
				}
				if lat != 0 && lon != 0 {
					s.SetSpeed(speedF)
					s.SetDirection(180)
				} else {
					s.SetSpeed(speedF)
				}
				s.SendLocation(s.CurLat(), s.CurLon(), uint16(speedF*10), 180)

			case "alarm":
				alarmType, _ := step.Params["type"].(string)
				var alarmFlag uint16
				switch alarmType {
				case "overspeed":
					alarmFlag = 0x0001
				case "fatigue":
					alarmFlag = 0x0002
				case "emergency":
					alarmFlag = 0x0004
				default:
					alarmFlag = 0x0001
				}
				if s, ok := sim.(*jt808sim.Simulator); ok {
					s.SendAlarm(alarmFlag)
				}

			case "wait":
				// 等待平台应答——已被 Delay 处理

			case "disconnect":
				sim.Offline()

			default:
				logger.Warn("场景未知action", "action", step.Action)
			}
			return nil
		})
		if err != nil {
			logger.Error("场景运行出错", "scenario", req.Name, "error", err)
		} else {
			logger.Info("场景运行完成", "scenario", req.Name)
		}
	}()

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "场景已启动: " + req.Name})
}

func (h *Handler) StopScenario(c *gin.Context) {
	if !h.scenarioEngine.IsRunning() {
		c.JSON(http.StatusOK, gin.H{"success": false, "error": "没有正在运行的场景"})
		return
	}
	h.scenarioEngine.Stop()
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "场景已停止"})
}

func (h *Handler) ScenarioStatus(c *gin.Context) {
	running := h.scenarioEngine.IsRunning()
	name := h.scenarioEngine.RunningName()
	current, total := h.scenarioEngine.Progress()

	resp := gin.H{
		"running":       running,
		"scenario_name": name,
		"current_step":  current,
		"total_steps":   total,
	}
	if total > 0 {
		resp["progress"] = fmt.Sprintf("%d/%d", current+1, total)
		resp["progress_percent"] = int(float64(current+1) / float64(total) * 100)
	} else {
		resp["progress"] = "-"
		resp["progress_percent"] = 0
	}
	c.JSON(http.StatusOK, resp)
}

// ============================== 压测 ==============================

func (h *Handler) RunStress(c *gin.Context) {
	var cfg engine.StressConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	if h.stressEngine.IsRunning() {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "压测正在运行中"})
		return
	}

	if cfg.DeviceCount <= 0 || cfg.DeviceCount > 10000 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "设备数量必须在 1-10000 之间"})
		return
	}
	if cfg.Duration <= 0 || cfg.Duration > 3600 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "持续时间必须在 1-3600 秒之间"})
		return
	}
	if cfg.TargetAddr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "目标地址不能为空"})
		return
	}
	if cfg.PhoneStart == "" {
		cfg.PhoneStart = "013800000000"
	}
	if cfg.Interval <= 0 {
		cfg.Interval = 30
	}

	go func() {
		result, err := h.stressEngine.Run(context.Background(), &cfg)
		if err != nil {
			logger.Error("压测运行错误", "error", err)
			return
		}
		logger.Info("压测完成", "online", result.OnlineDevices, "messages", result.TotalMessages, "errors", result.ErrorCount)
	}()

	c.JSON(http.StatusOK, gin.H{"success": true, "message": fmt.Sprintf("压测已启动: %d台设备, %d秒", cfg.DeviceCount, cfg.Duration)})
}

func (h *Handler) StopStress(c *gin.Context) {
	h.stressEngine.Stop()
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *Handler) StressStatus(c *gin.Context) {
	result := h.stressEngine.GetResult()
	resp := gin.H{
		"running": h.stressEngine.IsRunning(),
	}
	if result != nil {
		resp["result"] = result
	}
	c.JSON(http.StatusOK, resp)
}

// ============================== 统计 ==============================

func (h *Handler) GetStats(c *gin.Context) {
	stats := h.engine.GetStats()
	c.JSON(http.StatusOK, gin.H{
		"total_devices":  stats.TotalDevices,
		"online_devices": stats.OnlineDevices,
		"total_messages": stats.TotalMessages,
	})
}

// ============================== 健康检查 ==============================

func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
