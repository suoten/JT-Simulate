package engine

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/suoten/jt-simulate/internal/simulator/base"
	gbt32960sim "github.com/suoten/jt-simulate/internal/simulator/gbt32960"
	jt808sim "github.com/suoten/jt-simulate/internal/simulator/jt808"
	"github.com/suoten/jt-simulate/internal/storage"
	"github.com/suoten/jt-simulate/pkg/codec/jt808"
)

// BroadcastFunc 消息广播函数（用于WebSocket实时监控）
type BroadcastFunc func(direction, phone, msgName, msgID string, raw []byte)

// Simulator 接口：统一仿真器行为
type Simulator interface {
	Online(ctx context.Context) error
	Offline()
	Config() *base.DeviceConfig
	State() base.DeviceState
	SetOnState(func(base.DeviceState))
	SetOnSend(func([]byte))
	SetOnRawRecv(base.RawDataHandler)
}

// Engine 仿真引擎
type Engine struct {
	mu        sync.RWMutex
	devices   map[string]*DeviceInfo
	stats     EngineStats
	broadcast BroadcastFunc
	storage   *storage.Storage
}

// DeviceInfo 设备信息
type DeviceInfo struct {
	Config   *base.DeviceConfig
	Simulator interface{} // Simulator 接口实现
	State    base.DeviceState
}

// EngineStats 引擎统计
type EngineStats struct {
	TotalDevices  int64
	OnlineDevices int64
	TotalMessages int64
}

// New 创建引擎
func New() *Engine {
	return &Engine{
		devices: make(map[string]*DeviceInfo),
		storage: storage.New("data/devices.json"),
	}
}

// SetStorage 设置存储
func (e *Engine) SetStorage(s *storage.Storage) {
	e.storage = s
}

// LoadDevices 从持久化存储加载设备
func (e *Engine) LoadDevices() error {
	if e.storage == nil {
		return nil
	}
	devices, err := e.storage.LoadDevices()
	if err != nil {
		return fmt.Errorf("load devices: %w", err)
	}
	for _, cfg := range devices {
		sim, err := createSimulator(cfg)
		if err != nil {
			continue // 跳过无法创建的设备
		}
		e.mu.Lock()
		e.devices[cfg.ID] = &DeviceInfo{
			Config:    cfg,
			Simulator: sim,
			State:     base.StateOffline,
		}
		e.mu.Unlock()
		atomic.AddInt64(&e.stats.TotalDevices, 1)
	}
	return nil
}

// SaveDevices 保存所有设备到持久化存储
func (e *Engine) SaveDevices() error {
	if e.storage == nil {
		return nil
	}
	e.mu.RLock()
	defer e.mu.RUnlock()
	list := make([]*base.DeviceConfig, 0, len(e.devices))
	for _, info := range e.devices {
		list = append(list, info.Config)
	}
	return e.storage.SaveDevices(list)
}

// SetBroadcast 设置消息广播函数（由API层注入WebSocket Hub）
func (e *Engine) SetBroadcast(fn BroadcastFunc) {
	e.broadcast = fn
}

// createSimulator 根据协议创建仿真器
func createSimulator(cfg *base.DeviceConfig) (Simulator, error) {
	switch cfg.Protocol {
	case "jt808", "jt1078", "jt905", "jt1045", "jt1253":
		// 这些协议都使用 JT808 帧格式，JT808 仿真器即可处理
		return jt808sim.New(cfg), nil
	case "gbt32960":
		return gbt32960sim.New(cfg), nil
	case "jt809":
		// JT809 是平台间协议，使用 JT808 仿真器作为基础（平台间协议的仿真场景较少）
		return jt808sim.New(cfg), nil
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", cfg.Protocol)
	}
}

// CreateDevice 创建设备
func (e *Engine) CreateDevice(cfg *base.DeviceConfig) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, exists := e.devices[cfg.ID]; exists {
		return fmt.Errorf("device %s already exists", cfg.ID)
	}

	sim, err := createSimulator(cfg)
	if err != nil {
		return err
	}

	info := &DeviceInfo{
		Config:   cfg,
		Simulator: sim,
		State:    base.StateOffline,
	}
	e.devices[cfg.ID] = info
	atomic.AddInt64(&e.stats.TotalDevices, 1)

	// 持久化
	go e.SaveDevices()

	return nil
}

// StartDevice 启动设备
func (e *Engine) StartDevice(ctx context.Context, deviceID string) error {
	e.mu.RLock()
	info, ok := e.devices[deviceID]
	e.mu.RUnlock()

	if !ok {
		return fmt.Errorf("device %s not found", deviceID)
	}

	// 如果设备已经在线，不要重复启动
	if info.State == base.StateOnline || info.State == base.StateConnecting {
		return fmt.Errorf("device %s is already running", deviceID)
	}

	sim, ok := info.Simulator.(Simulator)
	if !ok {
		return fmt.Errorf("simulator does not implement Simulator interface")
	}

	sim.SetOnState(func(state base.DeviceState) {
		e.mu.Lock()
		info.State = state
		if state == base.StateOnline {
			atomic.AddInt64(&e.stats.OnlineDevices, 1)
		} else if state == base.StateOffline {
			// 只在之前是在线状态时才减1，避免负数
			cur := atomic.LoadInt64(&e.stats.OnlineDevices)
			if cur > 0 {
				atomic.AddInt64(&e.stats.OnlineDevices, -1)
			}
		}
		e.mu.Unlock()
	})
	// 设置发送回调——广播到WebSocket用于实时监控 + 计入消息统计
	sim.SetOnSend(func(data []byte) {
		atomic.AddInt64(&e.stats.TotalMessages, 1)
		if e.broadcast != nil {
			// 尝试解码消息名称
			codec := jt808.NewCodec()
			if msg, err := codec.Decode(data); err == nil {
				msgName := jt808.MsgName(msg.Header.MsgID)
				e.broadcast("up", info.Config.Phone, msgName,
					fmt.Sprintf("0x%04X", msg.Header.MsgID), data)
			} else {
				e.broadcast("up", info.Config.Phone, info.Config.Protocol, "", data)
			}
		}
	})
	// 设置接收回调——广播平台下发的消息
	sim.SetOnRawRecv(func(data []byte) {
		if e.broadcast != nil {
			codec := jt808.NewCodec()
			frames := jt808.SplitByDelimiter(data)
			if len(frames) == 0 {
				frames = [][]byte{data}
			}
			for _, frame := range frames {
				if msg, err := codec.Decode(frame); err == nil {
					msgName := jt808.MsgName(msg.Header.MsgID)
					e.broadcast("down", info.Config.Phone, msgName,
						fmt.Sprintf("0x%04X", msg.Header.MsgID), frame)
				} else {
					e.broadcast("down", info.Config.Phone, "未知", "", frame)
				}
			}
		}
	})
	return sim.Online(ctx)
}

// StopDevice 停止设备
func (e *Engine) StopDevice(deviceID string) error {
	e.mu.RLock()
	info, ok := e.devices[deviceID]
	e.mu.RUnlock()

	if !ok {
		return fmt.Errorf("device %s not found", deviceID)
	}

	// 如果设备已经离线，不需要重复停止
	if info.State == base.StateOffline {
		return nil
	}

	sim, ok := info.Simulator.(Simulator)
	if !ok {
		return fmt.Errorf("simulator does not implement Simulator interface")
	}
	sim.Offline()
	return nil
}

// DeleteDevice 删除设备
func (e *Engine) DeleteDevice(deviceID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	info, ok := e.devices[deviceID]
	if !ok {
		return fmt.Errorf("device %s not found", deviceID)
	}

	// 先停止
	if info.State != base.StateOffline {
		if sim, ok := info.Simulator.(Simulator); ok {
			sim.Offline()
		}
		// 减少在线设备计数
		cur := atomic.LoadInt64(&e.stats.OnlineDevices)
		if cur > 0 {
			atomic.AddInt64(&e.stats.OnlineDevices, -1)
		}
	}

	delete(e.devices, deviceID)
	atomic.AddInt64(&e.stats.TotalDevices, -1)

	// 持久化
	go e.SaveDevices()

	return nil
}

// GetDevice 获取设备
func (e *Engine) GetDevice(deviceID string) (*DeviceInfo, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	info, ok := e.devices[deviceID]
	if !ok {
		return nil, fmt.Errorf("device %s not found", deviceID)
	}
	return info, nil
}

// ListDevices 列出所有设备
func (e *Engine) ListDevices() []*DeviceInfo {
	e.mu.RLock()
	defer e.mu.RUnlock()

	list := make([]*DeviceInfo, 0, len(e.devices))
	for _, info := range e.devices {
		list = append(list, info)
	}
	return list
}

// GetStats 获取统计
func (e *Engine) GetStats() EngineStats {
	return EngineStats{
		TotalDevices:  atomic.LoadInt64(&e.stats.TotalDevices),
		OnlineDevices: atomic.LoadInt64(&e.stats.OnlineDevices),
		TotalMessages: atomic.LoadInt64(&e.stats.TotalMessages),
	}
}

// StopAll 停止所有设备
func (e *Engine) StopAll() {
	e.mu.RLock()
	defer e.mu.RUnlock()

	for _, info := range e.devices {
		if sim, ok := info.Simulator.(Simulator); ok {
			sim.Offline()
		}
	}
}
