package handler

import (
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/suoten/jt-simulate/internal/engine"
	"github.com/suoten/jt-simulate/internal/logger"
	"github.com/suoten/jt-simulate/internal/trajectory"
)

// maxTrajUploadSize 轨迹文件上传大小上限（20MB）
const maxTrajUploadSize = 20 << 20

// ============================== 轨迹回放 ==============================

// TrajectoryUpload 上传轨迹文件（multipart: file + 可选 name）
func (h *Handler) TrajectoryUpload(c *gin.Context) {
	if h.trajStore == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "轨迹存储未初始化"})
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "请选择轨迹文件（表单字段 file）"})
		return
	}
	if fileHeader.Size > maxTrajUploadSize {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "文件超过 20MB 大小限制"})
		return
	}

	name := strings.TrimSpace(c.PostForm("name"))
	if name == "" {
		// 默认取上传文件名（去扩展名）
		name = fileHeader.Filename
		if idx := strings.LastIndex(name, "."); idx > 0 {
			name = name[:idx]
		}
	}
	clean, err := trajectory.SanitizeName(name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	f, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "读取上传文件失败"})
		return
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, maxTrajUploadSize+1))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "读取上传文件失败"})
		return
	}
	if len(data) > maxTrajUploadSize {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "文件超过 20MB 大小限制"})
		return
	}

	route, err := trajectory.Parse(clean, data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	if err := h.trajStore.Save(route); err != nil {
		logger.Error("保存轨迹失败", "name", clean, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "保存轨迹失败: " + err.Error()})
		return
	}

	logger.Info("轨迹已导入", "name", route.Name, "points", len(route.Points),
		"distance_m", route.TotalDistM, "has_time", route.HasTime)
	c.JSON(http.StatusOK, gin.H{
		"success":          true,
		"name":             route.Name,
		"source":           route.Source,
		"point_count":      len(route.Points),
		"total_distance_m": route.TotalDistM,
		"duration_s":       route.DurationS,
		"has_time":         route.HasTime,
	})
}

// ListTrajectoryRoutes 轨迹列表
func (h *Handler) ListTrajectoryRoutes(c *gin.Context) {
	if h.trajStore == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "轨迹存储未初始化"})
		return
	}
	routes, err := h.trajStore.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	list := make([]gin.H, 0, len(routes))
	for _, r := range routes {
		list = append(list, trajSummary(r))
	}
	c.JSON(http.StatusOK, gin.H{"routes": list})
}

// GetTrajectoryRoute 轨迹详情（含全部点位）
func (h *Handler) GetTrajectoryRoute(c *gin.Context) {
	if h.trajStore == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "轨迹存储未初始化"})
		return
	}
	route, err := h.trajStore.Get(c.Param("name"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": err.Error()})
		return
	}
	resp := trajSummary(route)
	resp["points"] = route.Points
	c.JSON(http.StatusOK, resp)
}

// DeleteTrajectoryRoute 删除轨迹（回放中被占用时拒绝）
func (h *Handler) DeleteTrajectoryRoute(c *gin.Context) {
	if h.trajStore == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "轨迹存储未初始化"})
		return
	}
	name := c.Param("name")
	if h.trajEngine.UsesRoute(name) {
		c.JSON(http.StatusConflict, gin.H{"success": false, "error": "该轨迹正在回放中，请先停止回放"})
		return
	}
	if err := h.trajStore.Delete(name); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// StartTrajectory 开始回放：绑定指定终端与轨迹
func (h *Handler) StartTrajectory(c *gin.Context) {
	var opts engine.TrajOptions
	if err := c.ShouldBindJSON(&opts); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	if h.trajStore == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "轨迹存储未初始化"})
		return
	}
	route, err := h.trajStore.Get(opts.RouteName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	if err := h.trajEngine.Start(route, opts); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "轨迹回放已启动: " + route.Name})
}

// StopTrajectory 停止回放（device_id 为空时停止全部）
func (h *Handler) StopTrajectory(c *gin.Context) {
	var req struct {
		DeviceID string `json:"device_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil && !strings.Contains(err.Error(), "EOF") {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	if req.DeviceID == "" {
		h.trajEngine.StopAll()
		c.JSON(http.StatusOK, gin.H{"success": true, "message": "已停止全部轨迹回放"})
		return
	}
	if err := h.trajEngine.Stop(req.DeviceID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "轨迹回放已停止"})
}

// TrajectoryStatus 回放会话状态
func (h *Handler) TrajectoryStatus(c *gin.Context) {
	sessions := h.trajEngine.Status()
	c.JSON(http.StatusOK, gin.H{
		"running":  len(sessions) > 0,
		"sessions": sessions,
	})
}

func trajSummary(r *trajectory.Route) gin.H {
	return gin.H{
		"name":             r.Name,
		"source":           r.Source,
		"point_count":      len(r.Points),
		"total_distance_m": r.TotalDistM,
		"duration_s":       r.DurationS,
		"has_time":         r.HasTime,
	}
}
