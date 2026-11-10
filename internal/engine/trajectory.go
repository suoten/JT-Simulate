package engine

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/suoten/jt-simulate/internal/logger"
	"github.com/suoten/jt-simulate/internal/simulator/base"
	jt808sim "github.com/suoten/jt-simulate/internal/simulator/jt808"
	"github.com/suoten/jt-simulate/internal/trajectory"
)

// 回放模式
const (
	TrajModeUniform  = "uniform"  // 匀速推进：按配置车速沿轨迹等距推进
	TrajModeTimeline = "timeline" // 按时间回放：跟随轨迹文件中的时间轴（可倍速）
)

// jt808FamilyProtocols 支持 0x0200 位置上报的协议（复用 jt808 仿真器）
var jt808FamilyProtocols = map[string]bool{
	"jt808":  true,
	"jt1078": true,
	"jt905":  true,
	"jt1045": true,
	"jt1253": true,
}

// TrajOptions 轨迹回放配置
type TrajOptions struct {
	DeviceID    string  `json:"device_id"`
	RouteName   string  `json:"route_name"`
	IntervalSec int     `json:"interval"`     // 上报间隔（秒）
	Loop        bool    `json:"loop"`         // 循环回放
	Mode        string  `json:"mode"`         // uniform | timeline
	SpeedKmh    float64 `json:"speed"`        // uniform 模式模拟车速（km/h）
	SpeedFactor float64 `json:"speed_factor"` // timeline 模式倍速
}

// TrajSessionStatus 单个回放会话的对外状态
type TrajSessionStatus struct {
	DeviceID   string    `json:"device_id"`
	Phone      string    `json:"phone"`
	RouteName  string    `json:"route_name"`
	Mode       string    `json:"mode"`
	Interval   int       `json:"interval"`
	Loop       bool      `json:"loop"`
	Waiting    bool      `json:"waiting_device"` // 设备离线，等待恢复
	PointsSent int64     `json:"points_sent"`
	LoopsDone  int64     `json:"loops_done"`
	Progress   float64   `json:"progress_percent"` // 0-100
	CurLat     float64   `json:"cur_lat"`
	CurLon     float64   `json:"cur_lon"`
	CurSpeed   float64   `json:"cur_speed_kmh"`
	CurDir     uint16    `json:"cur_dir"`
	StartedAt  time.Time `json:"started_at"`
}

// trajSession 内部回放会话（推进状态由 run 循环持有，这里只存对外可见的进度）
type trajSession struct {
	opts  TrajOptions
	route *trajectory.Route
	sim   *jt808sim.Simulator
	phone string

	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{}

	mu         sync.Mutex
	waiting    bool
	pointsSent int64
	loops      int64
	progress   float64 // 0-100
	curLat     float64
	curLon     float64
	curSpeed   float64
	curDir     uint16
	startedAt  time.Time
}

func (s *trajSession) status() TrajSessionStatus {
	s.mu.Lock()
	defer s.mu.Unlock()
	return TrajSessionStatus{
		DeviceID:   s.opts.DeviceID,
		Phone:      s.phone,
		RouteName:  s.route.Name,
		Mode:       s.opts.Mode,
		Interval:   s.opts.IntervalSec,
		Loop:       s.opts.Loop,
		Waiting:    s.waiting,
		PointsSent: s.pointsSent,
		LoopsDone:  s.loops,
		Progress:   s.progress,
		CurLat:     s.curLat,
		CurLon:     s.curLon,
		CurSpeed:   s.curSpeed,
		CurDir:     s.curDir,
		StartedAt:  s.startedAt,
	}
}

func (s *trajSession) setWaiting(w bool) {
	s.mu.Lock()
	s.waiting = w
	s.mu.Unlock()
}

// record 上报成功后更新对外进度
func (s *trajSession) record(pos trajectory.Pos, speedKmh float64) {
	s.mu.Lock()
	s.pointsSent++
	s.progress = pos.Progress * 100
	s.curLat, s.curLon = pos.Lat, pos.Lon
	s.curSpeed, s.curDir = speedKmh, pos.Dir
	s.waiting = false
	s.mu.Unlock()
}

func (s *trajSession) addLoop() {
	s.mu.Lock()
	s.loops++
	s.mu.Unlock()
}

// TrajectoryEngine 轨迹回放引擎：管理每台设备的回放会话（每设备同时最多一个）。
// 回放期间接管设备的 0x0200 上报（SetPlaybackMode），按会话自身的间隔沿轨迹推进。
type TrajectoryEngine struct {
	mu        sync.Mutex
	sessions  map[string]*trajSession
	getDevice func(string) (*DeviceInfo, error)
}

// NewTrajectoryEngine 创建回放引擎，getDevice 用于设备查找与存活检测
func NewTrajectoryEngine(getDevice func(string) (*DeviceInfo, error)) *TrajectoryEngine {
	return &TrajectoryEngine{
		sessions:  make(map[string]*trajSession),
		getDevice: getDevice,
	}
}

// Start 为指定设备启动轨迹回放。同设备已有会话时先停止旧会话再启动新会话。
func (e *TrajectoryEngine) Start(route *trajectory.Route, opts TrajOptions) error {
	if route == nil {
		return fmt.Errorf("轨迹不存在")
	}
	if opts.DeviceID == "" {
		return fmt.Errorf("请指定要绑定轨迹的设备")
	}
	if opts.IntervalSec <= 0 || opts.IntervalSec > 3600 {
		opts.IntervalSec = 5
	}
	if opts.Mode == "" {
		opts.Mode = TrajModeUniform
	}
	switch opts.Mode {
	case TrajModeUniform:
		if opts.SpeedKmh <= 0 {
			opts.SpeedKmh = 40
		}
	case TrajModeTimeline:
		if !route.HasTime {
			return fmt.Errorf("轨迹 %q 不含时间信息，无法按时间回放（可改用匀速推进模式）", route.Name)
		}
		if opts.SpeedFactor <= 0 {
			opts.SpeedFactor = 1
		}
	default:
		return fmt.Errorf("未知回放模式: %s", opts.Mode)
	}

	info, err := e.getDevice(opts.DeviceID)
	if err != nil {
		return fmt.Errorf("设备不存在: %w", err)
	}
	if !jt808FamilyProtocols[info.Config.Protocol] {
		return fmt.Errorf("轨迹回放仅支持 JT808 系协议设备，当前协议: %s", info.Config.Protocol)
	}
	sim, ok := info.Simulator.(*jt808sim.Simulator)
	if !ok {
		return fmt.Errorf("设备仿真器类型不支持轨迹回放")
	}
	if info.State() != base.StateOnline {
		return fmt.Errorf("设备未在线，请先在设备管理中启动设备")
	}

	// 同设备已有会话：先停旧的
	if err := e.Stop(opts.DeviceID); err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	s := &trajSession{
		opts:      opts,
		route:     route,
		sim:       sim,
		phone:     info.Config.Phone,
		ctx:       ctx,
		cancel:    cancel,
		done:      make(chan struct{}),
		startedAt: time.Now(),
	}

	e.mu.Lock()
	e.sessions[opts.DeviceID] = s
	e.mu.Unlock()

	go e.run(s)

	logger.Info("轨迹回放已启动", "device_id", opts.DeviceID, "route", route.Name,
		"mode", opts.Mode, "interval", opts.IntervalSec, "loop", opts.Loop)
	return nil
}

// Stop 停止指定设备的回放会话（无会话时幂等返回）
func (e *TrajectoryEngine) Stop(deviceID string) error {
	e.mu.Lock()
	s, ok := e.sessions[deviceID]
	if ok {
		delete(e.sessions, deviceID)
	}
	e.mu.Unlock()
	if !ok {
		return nil
	}
	s.cancel()
	select {
	case <-s.done:
	case <-time.After(3 * time.Second):
		logger.Warn("轨迹回放会话退出超时", "device_id", deviceID)
	}
	logger.Info("轨迹回放已停止", "device_id", deviceID)
	return nil
}

// StopAll 停止所有回放会话（服务关闭时调用）
func (e *TrajectoryEngine) StopAll() {
	e.mu.Lock()
	ids := make([]string, 0, len(e.sessions))
	for id := range e.sessions {
		ids = append(ids, id)
	}
	e.mu.Unlock()
	for _, id := range ids {
		_ = e.Stop(id)
	}
}

// Status 列出所有运行中的回放会话状态
func (e *TrajectoryEngine) Status() []TrajSessionStatus {
	e.mu.Lock()
	list := make([]*trajSession, 0, len(e.sessions))
	for _, s := range e.sessions {
		list = append(list, s)
	}
	e.mu.Unlock()

	out := make([]TrajSessionStatus, 0, len(list))
	for _, s := range list {
		out = append(out, s.status())
	}
	return out
}

// RunningCount 运行中的会话数
func (e *TrajectoryEngine) RunningCount() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return len(e.sessions)
}

// UsesRoute 是否有会话正在回放指定轨迹（删除轨迹前的占用检查）
func (e *TrajectoryEngine) UsesRoute(name string) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	for _, s := range e.sessions {
		if s.route.Name == name {
			return true
		}
	}
	return false
}

// removeSession 会话自行结束时从会话表移除（Stop 主动取消时该表项已被删，此处为幂等兜底）
func (e *TrajectoryEngine) removeSession(s *trajSession) {
	e.mu.Lock()
	if cur, ok := e.sessions[s.opts.DeviceID]; ok && cur == s {
		delete(e.sessions, s.opts.DeviceID)
	}
	e.mu.Unlock()
}

// run 回放主循环
func (e *TrajectoryEngine) run(s *trajSession) {
	defer close(s.done)
	defer e.removeSession(s)

	// 接管期间抑制设备的默认周期位置上报；会话结束（含取消/回放完毕）后
	// 从轨迹当前位置恢复默认上报
	s.sim.SetPlaybackMode(true)
	defer s.sim.SetPlaybackMode(false)

	ticker := time.NewTicker(time.Duration(s.opts.IntervalSec) * time.Second)
	defer ticker.Stop()

	// sendAt 推进到 pos 并上报 0x0200。fallbackSpeed 用于轨迹未提供速度时的回退。
	sendAt := func(pos trajectory.Pos, fallbackSpeed float64) bool {
		speed := pos.SpeedKmh
		if speed <= 0 {
			speed = fallbackSpeed
		}
		if speed > 500 { // 合理车速上限，避免异常文件产生荒谬值
			speed = 500
		}
		dir := pos.Dir
		if dir > 359 {
			dir = 359
		}
		// 同步更新仿真器运行时位置：0x8201 位置查询应答与回放保持一致
		s.sim.SetPosition(pos.Lat, pos.Lon)
		s.sim.SetSpeed(speed)
		s.sim.SetDirection(dir)
		if err := s.sim.SendLocation(pos.Lat, pos.Lon, uint16(speed*10), dir); err != nil {
			// 发送失败多为设备正在断线：标记等待，等重连后从当前进度继续
			s.setWaiting(true)
			return false
		}
		s.record(pos, speed)
		return true
	}

	var (
		dist  float64 // uniform 模式：沿轨迹的累计距离（米）
		clock float64 // timeline 模式：回放时钟（相对轨迹起点的秒数）
	)

	// 启动即上报轨迹起点
	if s.opts.Mode == TrajModeTimeline {
		sendAt(s.route.PositionAtTime(0), 0)
	} else {
		sendAt(s.route.PositionAtDist(0), s.opts.SpeedKmh)
	}

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			// 设备已被删除则结束会话
			info, err := e.getDevice(s.opts.DeviceID)
			if err != nil {
				logger.Info("设备已删除，轨迹回放结束", "device_id", s.opts.DeviceID, "route", s.route.Name)
				return
			}
			if info.State() != base.StateOnline {
				// 设备离线：暂停推进（进度保持），重连后从当前位置继续
				s.setWaiting(true)
				continue
			}

			if s.opts.Mode == TrajModeTimeline {
				clock += float64(s.opts.IntervalSec) * s.opts.SpeedFactor
				if clock >= s.route.DurationS {
					if !s.opts.Loop {
						sendAt(s.route.PositionAtTime(s.route.DurationS), 0)
						return
					}
					// 循环回放：回绕并计圈（单 tick 可能跨多圈，理论上不会，防御处理）
					for clock >= s.route.DurationS {
						clock -= s.route.DurationS
						s.addLoop()
					}
				}
				sendAt(s.route.PositionAtTime(clock), 0)
			} else {
				dist += s.opts.SpeedKmh / 3.6 * float64(s.opts.IntervalSec)
				if dist >= s.route.TotalDistM {
					if !s.opts.Loop {
						sendAt(s.route.PositionAtDist(s.route.TotalDistM), s.opts.SpeedKmh)
						return
					}
					for dist >= s.route.TotalDistM && s.route.TotalDistM > 0 {
						dist -= s.route.TotalDistM
						s.addLoop()
					}
				}
				sendAt(s.route.PositionAtDist(dist), s.opts.SpeedKmh)
			}
		}
	}
}
