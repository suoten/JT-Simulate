package engine

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/suoten/jt-simulate/internal/simulator/base"
	jt808sim "github.com/suoten/jt-simulate/internal/simulator/jt808"
)

// StressConfig 压测配置
type StressConfig struct {
	DeviceCount  int    `json:"device_count"`
	Duration     int    `json:"duration"` // 秒
	TargetAddr   string `json:"target_addr"`
	Protocol     string `json:"protocol"`
	PhoneStart   string `json:"phone_start"`
	Interval     int    `json:"interval"` // 位置上报间隔（秒）
}

// StressResult 压测结果
type StressResult struct {
	TotalDevices  int64         `json:"total_devices"`
	OnlineDevices int64         `json:"online_devices"`
	TotalMessages int64         `json:"total_messages"`
	ErrorCount    int64         `json:"error_count"`
	Duration      int           `json:"duration"`
	AvgMsgRate    float64       `json:"avg_msg_rate"`  // 消息/秒
	AvgLatency    float64       `json:"avg_latency"`  // 毫秒
	StartTime     time.Time     `json:"start_time"`
	EndTime       time.Time     `json:"end_time"`
	Errors        []string      `json:"errors,omitempty"`
}

// StressEngine 压测引擎
type StressEngine struct {
	mu      sync.RWMutex
	running bool
	cancel  context.CancelFunc
	result  *StressResult
	msgCount  int64
	errCount  int64
}

// NewStressEngine 创建压测引擎
func NewStressEngine() *StressEngine {
	return &StressEngine{}
}

// Run 运行压测
func (e *StressEngine) Run(ctx context.Context, cfg *StressConfig) (*StressResult, error) {
	ctx, cancel := context.WithCancel(ctx)
	e.mu.Lock()
	e.running = true
	e.cancel = cancel
	e.msgCount = 0
	e.errCount = 0
	e.mu.Unlock()

	defer func() {
		e.mu.Lock()
		e.running = false
		e.mu.Unlock()
	}()

	result := &StressResult{
		StartTime: time.Now(),
		TotalDevices: int64(cfg.DeviceCount),
	}
	e.result = result

	// 按批次创建设备
	batchSize := 50
	if cfg.DeviceCount < batchSize {
		batchSize = cfg.DeviceCount
	}

	var wg sync.WaitGroup
	var devices []*jt808sim.Simulator

	for i := 0; i < cfg.DeviceCount; i++ {
		select {
		case <-ctx.Done():
			goto waitForDevices
		case <-time.After(10 * time.Millisecond): // 批量启动间隔
		}

		phone := incrementPhone(cfg.PhoneStart, i)
		devCfg := &base.DeviceConfig{
			ID:         phone,
			Protocol:   cfg.Protocol,
			Phone:      phone,
			Plate:      fmt.Sprintf("测%05d", i),
			PlateColor: 2,
			TargetAddr: cfg.TargetAddr,
			AuthCode:   "stress_test",
			ProvinceID: 11,
			CityID:     100,
			Manufacturer:  "STRESS",
			TerminalModel: "S-100",
			TerminalID:    fmt.Sprintf("%07d", i),
			HeartbeatInterval: 60,
			LocationInterval:  cfg.Interval,
			ReconnectInterval: 15,
			StartLat: 39.9093 + float64(i)*0.0001,
			StartLon: 116.3974 + float64(i)*0.0001,
		}

		sim := jt808sim.New(devCfg)
		devices = append(devices, sim)

		wg.Add(1)
		go func(s *jt808sim.Simulator, idx int) {
			defer wg.Done()
			if err := s.Online(ctx); err != nil {
				atomic.AddInt64(&e.errCount, 1)
				return
			}
			atomic.AddInt64(&e.msgCount, 1) // 注册消息
			atomic.AddInt64(&result.OnlineDevices, 1)
		}(sim, i)

		// 批量等待
		if (i+1)%batchSize == 0 {
			time.Sleep(100 * time.Millisecond)
		}
	}

	// 等待所有设备尝试连接
waitForDevices:
	wg.Wait()

	// 等待压测持续时间
	timer := time.NewTimer(time.Duration(cfg.Duration) * time.Second)
	defer timer.Stop()

	select {
	case <-ctx.Done():
	case <-timer.C:
	}

	// 停止所有设备
	for _, sim := range devices {
		sim.Offline()
	}

	result.EndTime = time.Now()
	result.Duration = int(result.EndTime.Sub(result.StartTime).Seconds())
	result.TotalMessages = atomic.LoadInt64(&e.msgCount)
	result.ErrorCount = atomic.LoadInt64(&e.errCount)
	result.OnlineDevices = atomic.LoadInt64(&result.OnlineDevices)
	if result.Duration > 0 {
		result.AvgMsgRate = float64(result.TotalMessages) / float64(result.Duration)
	}

	return result, nil
}

// Stop 停止压测
func (e *StressEngine) Stop() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.cancel != nil {
		e.cancel()
	}
	e.running = false
}

// IsRunning 是否正在运行
func (e *StressEngine) IsRunning() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.running
}

// GetResult 获取压测结果（运行中返回实时进度，结束后返回最终结果）
func (e *StressEngine) GetResult() *StressResult {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if e.result == nil {
		return nil
	}
	// 返回实时或最终结果
	r := &StressResult{
		TotalDevices:  e.result.TotalDevices,
		OnlineDevices: atomic.LoadInt64(&e.result.OnlineDevices),
		TotalMessages: atomic.LoadInt64(&e.msgCount),
		ErrorCount:    atomic.LoadInt64(&e.errCount),
		StartTime:     e.result.StartTime,
	}
	if e.running {
		r.Duration = int(time.Since(e.result.StartTime).Seconds())
	} else {
		r.Duration = e.result.Duration
		r.EndTime = e.result.EndTime
		r.AvgLatency = e.result.AvgLatency
		r.Errors = e.result.Errors
	}
	if r.Duration > 0 {
		r.AvgMsgRate = float64(r.TotalMessages) / float64(r.Duration)
	}
	return r
}

// incrementPhone 手机号递增
func incrementPhone(base string, inc int) string {
	if len(base) != 12 {
		return fmt.Sprintf("0138000%05d", inc)
	}
	// 取后10位数字递增
	result := []byte(base)
	carry := inc
	for i := 11; i >= 2 && carry > 0; i-- {
		if result[i] < '0' || result[i] > '9' {
			continue
		}
		d := int(result[i]-'0') + carry
		result[i] = byte('0' + d%10)
		carry = d / 10
	}
	return string(result)
}
