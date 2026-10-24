package engine

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/suoten/jt-simulate/internal/logger"
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
	// TrySetState 原子状态迁移，用于并发启动保护
	TrySetState(from, to base.DeviceState) bool
	SetOnState(func(from, to base.DeviceState))
	SetOnSend(func([]byte))
	SetOnRawRecv(base.RawDataHandler)
	WrapOnRawRecv(base.RawDataHandler)
}

// Engine 仿真引擎
type Engine struct {
	mu        sync.RWMutex
	devices   map[string]*DeviceInfo
	stats     EngineStats
	broadcast BroadcastFunc
	storage   storage.Storage

	// rootCtx 所有设备会话的父 context：设备不随单个 HTTP 请求结束而取消，
	// 由 StopAll/进程关闭统一取消
	rootCtx    context.Context
	rootCancel context.CancelFunc
}

// DeviceInfo 设备信息
type DeviceInfo struct {
	Config    *base.DeviceConfig
	Simulator Simulator // 直接存储 Simulator 接口，避免反复类型断言
	state     atomic.Int32
}

// State 获取设备状态（线程安全）
func (d *DeviceInfo) State() base.DeviceState {
	return base.DeviceState(d.state.Load())
}

// setState 设置设备状态（线程安全）
func (d *DeviceInfo) setState(s base.DeviceState) {
	d.state.Store(int32(s))
}

// EngineStats 引擎统计
type EngineStats struct {
	TotalDevices  int64
	OnlineDevices int64
	TotalMessages int64
}

// New 创建引擎
func New() *Engine {
	s := storage.New("data/devices.json")
	rootCtx, rootCancel := context.WithCancel(context.Background())
	return &Engine{
		devices:    make(map[string]*DeviceInfo),
		storage:    storage.NewDebounced(s, 2*time.Second),
		rootCtx:    rootCtx,
		rootCancel: rootCancel,
	}
}

// SetStorage 设置存储
func (e *Engine) SetStorage(s storage.Storage) {
	if s == nil {
		e.storage = storage.NoopStorage()
	} else {
		e.storage = s
	}
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
			logger.Warn("跳过无法创建的设备", "device_id", cfg.ID, "error", err)
			continue
		}
		e.mu.Lock()
		info := &DeviceInfo{
			Config:    cfg,
			Simulator: sim,
		}
		info.setState(base.StateOffline)
		e.devices[cfg.ID] = info
		e.mu.Unlock()
		e.attachCallbacks(cfg.ID, info)
		atomic.AddInt64(&e.stats.TotalDevices, 1)
	}
	if len(devices) > 0 {
		logger.Info("从持久化存储加载设备", "count", len(devices))
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
	if err := e.storage.SaveDevices(list); err != nil {
		logger.Error("保存设备列表失败", "error", err)
		return err
	}
	return nil
}

// SetBroadcast 设置消息广播函数（由API层注入WebSocket Hub）
func (e *Engine) SetBroadcast(fn BroadcastFunc) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.broadcast = fn
}

// createSimulator 根据协议创建仿真器
func createSimulator(cfg *base.DeviceConfig) (Simulator, error) {
	switch cfg.Protocol {
	case "jt808", "jt1078", "jt905", "jt1045", "jt1253":
		return jt808sim.New(cfg), nil
	case "gbt32960":
		return gbt32960sim.New(cfg), nil
	case "jt809":
		return jt808sim.New(cfg), nil
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", cfg.Protocol)
	}
}

// CreateDevice 创建设备
func (e *Engine) CreateDevice(cfg *base.DeviceConfig) error {
	e.mu.Lock()
	if _, exists := e.devices[cfg.ID]; exists {
		e.mu.Unlock()
		return fmt.Errorf("device %s already exists", cfg.ID)
	}

	sim, err := createSimulator(cfg)
	if err != nil {
		e.mu.Unlock()
		return err
	}

	info := &DeviceInfo{
		Config:    cfg,
		Simulator: sim,
	}
	info.setState(base.StateOffline)
	e.devices[cfg.ID] = info
	e.mu.Unlock()

	e.attachCallbacks(cfg.ID, info)

	atomic.AddInt64(&e.stats.TotalDevices, 1)

	go e.SaveDevices()

	logger.Info("设备已创建", "device_id", cfg.ID, "protocol", cfg.Protocol)
	return nil
}

// attachCallbacks 为设备挂载引擎回调（在设备创建/加载时调用一次，切勿在每次启动时
// 重复调用——WrapOnRawRecv 是追加语义，重复包装会导致下行帧被广播 N 次）
func (e *Engine) attachCallbacks(deviceID string, info *DeviceInfo) {
	sim := info.Simulator
	// 缓存 phone/protocol 避免在回调闭包中反复访问 info.Config
	phone := info.Config.Phone
	protocol := info.Config.Protocol

	sim.SetOnState(func(from, to base.DeviceState) {
		info.setState(to)
		// 边沿触发计数：只有 非Online→Online / Online→非Online 的迁移才增减在线数，
		// 同一事件（如 Disconnect 的 Closing→Offline 与回调末尾的 Offline）不会重复扣减
		if from != base.StateOnline && to == base.StateOnline {
			atomic.AddInt64(&e.stats.OnlineDevices, 1)
		} else if from == base.StateOnline && to != base.StateOnline {
			cur := atomic.LoadInt64(&e.stats.OnlineDevices)
			if cur > 0 {
				atomic.AddInt64(&e.stats.OnlineDevices, -1)
			}
		}
		logger.Debug("设备状态变更", "device_id", deviceID, "from", from.String(), "to", to.String())
	})

	sim.SetOnSend(func(data []byte) {
		atomic.AddInt64(&e.stats.TotalMessages, 1)
		e.mu.RLock()
		bc := e.broadcast
		e.mu.RUnlock()
		if bc != nil {
			codec := jt808.NewCodec()
			if msg, err := codec.Decode(data); err == nil {
				msgName := jt808.MsgName(msg.Header.MsgID)
				bc("up", phone, msgName,
					fmt.Sprintf("0x%04X", msg.Header.MsgID), data)
			} else {
				bc("up", phone, protocol, "", data)
			}
		}
	})

	sim.WrapOnRawRecv(func(data []byte) {
		e.mu.RLock()
		bc := e.broadcast
		e.mu.RUnlock()
		if bc != nil {
			codec := jt808.NewCodec()
			frames := jt808.SplitByDelimiter(data)
			if len(frames) == 0 {
				frames = [][]byte{data}
			}
			for _, frame := range frames {
				if msg, err := codec.Decode(frame); err == nil {
					msgName := jt808.MsgName(msg.Header.MsgID)
					bc("down", phone, msgName,
						fmt.Sprintf("0x%04X", msg.Header.MsgID), frame)
				} else {
					bc("down", phone, "未知", "", frame)
				}
			}
		}
	})
}

// StartDevice 启动设备
// ctx 参数仅为接口兼容保留：设备实际挂在引擎 root ctx 下（不随 HTTP 请求结束而取消，
// 避免心跳/位置上报在响应返回后被取消；引擎 StopAll 时统一停止）。
func (e *Engine) StartDevice(ctx context.Context, deviceID string) error {
	e.mu.RLock()
	info, ok := e.devices[deviceID]
	e.mu.RUnlock()

	if !ok {
		return fmt.Errorf("device %s not found", deviceID)
	}

	if info.State() == base.StateOnline || info.State() == base.StateConnecting {
		return fmt.Errorf("device %s is already running", deviceID)
	}

	logger.Info("启动设备", "device_id", deviceID, "target", info.Config.TargetAddr)
	e.mu.RLock()
	rootCtx := e.rootCtx
	e.mu.RUnlock()
	return info.Simulator.Online(rootCtx)
}

// StopDevice 停止设备
func (e *Engine) StopDevice(deviceID string) error {
	e.mu.RLock()
	info, ok := e.devices[deviceID]
	e.mu.RUnlock()

	if !ok {
		return fmt.Errorf("device %s not found", deviceID)
	}

	if info.State() == base.StateOffline {
		return nil
	}

	sim := info.Simulator
	sim.Offline()
	logger.Info("停止设备", "device_id", deviceID)
	return nil
}

// DeleteDevice 删除设备
func (e *Engine) DeleteDevice(deviceID string) error {
	e.mu.Lock()
	info, ok := e.devices[deviceID]
	if !ok {
		e.mu.Unlock()
		return fmt.Errorf("device %s not found", deviceID)
	}

	// 在持锁状态下取出 sim，然后释放锁再调用 Offline（避免死锁）
	sim := info.Simulator
	wasOnline := info.State() != base.StateOffline
	delete(e.devices, deviceID)
	e.mu.Unlock()

	if wasOnline {
		// 在线数扣减由 onState 边沿回调完成，这里不再手动重复扣
		sim.Offline()
	}
	atomic.AddInt64(&e.stats.TotalDevices, -1)

	go e.SaveDevices()

	logger.Info("删除设备", "device_id", deviceID)
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

// StopAll 停止所有设备（优雅关闭）
// 注意：不持锁调用 Offline，因为 Offline 内部会 Wait goroutine，
// 而 goroutine 中的回调可能需要获取锁，导致死锁。
func (e *Engine) StopAll() {
	e.mu.RLock()
	type simEntry struct {
		sim    Simulator
		online bool
	}
	entries := make([]simEntry, 0, len(e.devices))
	for _, info := range e.devices {
		entries = append(entries, simEntry{
			sim:    info.Simulator,
			online: info.State() != base.StateOffline,
		})
	}
	e.mu.RUnlock()

	// 先取消引擎 root ctx：所有设备会话（含重连循环）随之开始退出，
	// 之后的 Offline 只需等待收尾，避免等待期间设备再次重连
	e.mu.Lock()
	if e.rootCancel != nil {
		e.rootCancel()
	}
	e.rootCtx, e.rootCancel = context.WithCancel(context.Background())
	e.mu.Unlock()

	// 释放锁后再停止设备，避免死锁
	count := 0
	for _, entry := range entries {
		if entry.online {
			entry.sim.Offline()
			count++
		}
	}
	if count > 0 {
		logger.Info("已停止所有在线设备", "count", count)
	}

	// 保存设备列表（如果是防抖存储，确保挂起的保存被刷新）
	if e.storage != nil {
		e.mu.RLock()
		list := make([]*base.DeviceConfig, 0, len(e.devices))
		for _, info := range e.devices {
			list = append(list, info.Config)
		}
		e.mu.RUnlock()
		if err := e.storage.SaveDevices(list); err != nil {
			logger.Error("关闭时保存设备列表失败", "error", err)
		}
		// 如果是防抖存储，刷新挂起的保存
		if ds, ok := e.storage.(interface{ Flush() error }); ok {
			if err := ds.Flush(); err != nil {
				logger.Error("关闭时刷新存储失败", "error", err)
			}
		}
	}
}
