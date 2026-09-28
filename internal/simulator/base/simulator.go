package base

import (
	"context"
	"fmt"
	"math"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/suoten/jt-simulate/internal/logger"
	"github.com/suoten/jt-simulate/pkg/types"
)

// DeviceState 设备状态
type DeviceState int32

const (
	StateOffline    DeviceState = 0
	StateConnecting DeviceState = 1
	StateOnline     DeviceState = 2
	StateClosing    DeviceState = 3
)

func (s DeviceState) String() string {
	switch s {
	case StateOffline:
		return "offline"
	case StateConnecting:
		return "connecting"
	case StateOnline:
		return "online"
	case StateClosing:
		return "closing"
	default:
		return "unknown"
	}
}

// DeviceConfig 设备配置（不可变部分，JSON 序列化）
type DeviceConfig struct {
	ID         string `json:"id"`
	Protocol   string `json:"protocol"`
	Phone      string `json:"phone"`
	Plate      string `json:"plate"`
	PlateColor byte   `json:"plate_color"`
	TargetAddr string `json:"target_addr"`
	AuthCode   string `json:"auth_code"`

	ProvinceID    int    `json:"province_id"`
	CityID        int    `json:"city_id"`
	Manufacturer  string `json:"manufacturer"`
	TerminalModel string `json:"terminal_model"`
	TerminalID    string `json:"terminal_id"`
	IMEI          string `json:"imei"`

	HeartbeatInterval int `json:"heartbeat_interval"`
	LocationInterval  int `json:"location_interval"`
	ReconnectInterval int `json:"reconnect_interval"`

	Route   string  `json:"route"`
	StartLat float64 `json:"start_lat"`
	StartLon float64 `json:"start_lon"`
}

// runtimeState 运行时状态（可变，需要加锁保护）
type runtimeState struct {
	mu        sync.Mutex
	curLat    float64
	curLon    float64
	curSpeed  float64 // km/h
	curDir    uint16  // 0-359
	authCode  string  // 运行时从平台收到的鉴权码
}

// MessageHandler 消息处理器
type MessageHandler func(msg *types.Message)

// RawDataHandler 原始数据处理器（子类设置后，recvLoop会把收到的原始数据传给它）
type RawDataHandler func(data []byte)

// Simulator 仿真器基类
type Simulator struct {
	config     *DeviceConfig
	rt         runtimeState // 运行时状态（带锁）
	state      atomic.Int32
	conn       net.Conn
	connMu     sync.RWMutex
	seqNum     uint16
	seqMu      sync.Mutex
	cancel     context.CancelFunc
	ctx        context.Context
	wg         sync.WaitGroup
	onRecv     MessageHandler
	onRawRecv  RawDataHandler
	onState    func(DeviceState)
	onSend     func([]byte)
	mu         sync.RWMutex
}

// NewSimulator 创建仿真器
func NewSimulator(cfg *DeviceConfig) *Simulator {
	if cfg.HeartbeatInterval <= 0 {
		cfg.HeartbeatInterval = 60
	}
	if cfg.LocationInterval <= 0 {
		cfg.LocationInterval = 30
	}
	if cfg.ReconnectInterval <= 0 {
		cfg.ReconnectInterval = 15
	}
	s := &Simulator{
		config: cfg,
	}
	s.rt.curLat = cfg.StartLat
	s.rt.curLon = cfg.StartLon
	if s.rt.curLat == 0 {
		s.rt.curLat = 39.9093
	}
	if s.rt.curLon == 0 {
		s.rt.curLon = 116.3974
	}
	return s
}

// Config 获取配置
func (s *Simulator) Config() *DeviceConfig {
	return s.config
}

// State 获取状态
func (s *Simulator) State() DeviceState {
	return DeviceState(s.state.Load())
}

// SetState 设置状态
func (s *Simulator) SetState(state DeviceState) {
	s.state.Store(int32(state))
	s.mu.RLock()
	fn := s.onState
	s.mu.RUnlock()
	if fn != nil {
		fn(state)
	}
}

// SetOnRecv 设置接收回调
func (s *Simulator) SetOnRecv(handler MessageHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onRecv = handler
}

// SetOnState 设置状态变更回调
func (s *Simulator) SetOnState(handler func(DeviceState)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onState = handler
}

// SetOnSend 设置发送回调
func (s *Simulator) SetOnSend(handler func([]byte)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onSend = handler
}

// SetOnRawRecv 设置原始数据接收回调
func (s *Simulator) SetOnRawRecv(handler RawDataHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onRawRecv = handler
}

// WrapOnRawRecv 在现有的 onRawRecv 回调之外追加一个回调（不覆盖子类在 New 中设置的 handler）
// 新 handler 会在原有 handler 之后被调用
func (s *Simulator) WrapOnRawRecv(extra RawDataHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	existing := s.onRawRecv
	if existing == nil {
		s.onRawRecv = extra
		return
	}
	s.onRawRecv = func(data []byte) {
		existing(data) // 先调用子类的消息处理（如自动应答）
		if extra != nil {
			extra(data) // 再调用引擎的广播
		}
	}
}

// NextSeqNum 获取下一个序号
func (s *Simulator) NextSeqNum() uint16 {
	s.seqMu.Lock()
	defer s.seqMu.Unlock()
	s.seqNum++
	return s.seqNum
}

// Connect 连接目标平台（带重连）
func (s *Simulator) Connect(ctx context.Context) error {
	s.ctx = ctx
	return s.connectAndRun(ctx)
}

// connectAndRun 连接并启动接收循环
func (s *Simulator) connectAndRun(ctx context.Context) error {
	s.SetState(StateConnecting)

	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", s.config.TargetAddr)
	if err != nil {
		s.SetState(StateOffline)
		return fmt.Errorf("connect %s: %w", s.config.TargetAddr, err)
	}

	s.connMu.Lock()
	s.conn = conn
	s.connMu.Unlock()
	s.SetState(StateOnline)

	// 启动接收循环
	ctx, s.cancel = context.WithCancel(ctx)
	s.ctx = ctx
	s.wg.Add(1)
	go s.recvLoop(ctx)

	return nil
}

// Context 返回内部可取消的 context（供子类在启动 goroutine 时使用）
func (s *Simulator) Context() context.Context {
	return s.ctx
}

// StartReconnectLoop 启动自动重连循环
func (s *Simulator) StartReconnectLoop(ctx context.Context, onReconnect func(ctx context.Context) error) {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			if s.State() == StateOffline {
				interval := s.config.ReconnectInterval
				if interval <= 0 {
					interval = 15
				}
				select {
				case <-ctx.Done():
					return
				case <-time.After(time.Duration(interval) * time.Second):
				}

				if s.State() != StateOffline {
					continue
				}
				logger.Info("尝试重连", "phone", s.config.Phone, "target", s.config.TargetAddr)
				if err := s.connectAndRun(ctx); err != nil {
					logger.Error("重连失败", "phone", s.config.Phone, "error", err)
					continue
				}
				if onReconnect != nil {
					if err := onReconnect(ctx); err != nil {
						logger.Error("重连后重新上线失败", "phone", s.config.Phone, "error", err)
						s.Disconnect()
						continue
					}
				}
				logger.Info("重连成功", "phone", s.config.Phone)
			} else {
				select {
				case <-ctx.Done():
					return
				case <-time.After(5 * time.Second):
				}
			}
		}
	}()
}

// Disconnect 断开连接
func (s *Simulator) Disconnect() {
	s.SetState(StateClosing)

	if s.cancel != nil {
		s.cancel()
	}

	s.connMu.Lock()
	if s.conn != nil {
		s.conn.Close()
		s.conn = nil
	}
	s.connMu.Unlock()

	s.wg.Wait()
	s.SetState(StateOffline)
}

// Send 发送原始数据
func (s *Simulator) Send(data []byte) error {
	s.connMu.RLock()
	conn := s.conn
	s.connMu.RUnlock()

	if conn == nil {
		return fmt.Errorf("not connected")
	}

	_, err := conn.Write(data)
	if err != nil {
		return fmt.Errorf("send: %w", err)
	}

	s.mu.RLock()
	fn := s.onSend
	s.mu.RUnlock()
	if fn != nil {
		fn(data)
	}

	return nil
}

// recvLoop 接收循环
func (s *Simulator) recvLoop(ctx context.Context) {
	defer s.wg.Done()

	buf := make([]byte, 4096)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		s.connMu.RLock()
		conn := s.conn
		s.connMu.RUnlock()
		if conn == nil {
			return
		}

		conn.SetReadDeadline(time.Now().Add(30 * time.Second))
		n, err := conn.Read(buf)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			// 连接断开
			s.SetState(StateOffline)
			return
		}

		if n > 0 {
			data := make([]byte, n)
			copy(data, buf[:n])
			// 传递原始数据给子类处理
			s.mu.RLock()
			handler := s.onRawRecv
			s.mu.RUnlock()
			if handler != nil {
				handler(data)
			}
		}
	}
}

// StartHeartbeat 启动心跳
func (s *Simulator) StartHeartbeat(ctx context.Context, sendFunc func() error) {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(time.Duration(s.config.HeartbeatInterval) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if s.State() != StateOnline {
					continue
				}
				if err := sendFunc(); err != nil {
					return
				}
			}
		}
	}()
}

// StartLocationReport 启动定时位置上报
func (s *Simulator) StartLocationReport(ctx context.Context, reportFunc func() error) {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(time.Duration(s.config.LocationInterval) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if s.State() != StateOnline {
					continue
				}
				if err := reportFunc(); err != nil {
					return
				}
			}
		}
	}()
}

// MoveAlongRoute 沿路线移动一个步长，返回新的经纬度
func (s *Simulator) MoveAlongRoute() (lat, lon float64) {
	s.rt.mu.Lock()
	defer s.rt.mu.Unlock()

	// 简单模拟：沿当前方向移动一定距离
	interval := float64(s.config.LocationInterval)
	distance := s.rt.curSpeed * 1000 / 3600 * interval // 米
	// 1度纬度 ≈ 111000米
	latDelta := distance / 111000 * math.Cos(float64(s.rt.curDir)*math.Pi/180)
	lonDelta := distance / (111000 * math.Cos(s.rt.curLat*math.Pi/180)) * math.Sin(float64(s.rt.curDir)*math.Pi/180)
	s.rt.curLat += latDelta
	s.rt.curLon += lonDelta
	return s.rt.curLat, s.rt.curLon
}

// SetSpeed 设置当前速度
func (s *Simulator) SetSpeed(kmh float64) {
	s.rt.mu.Lock()
	defer s.rt.mu.Unlock()
	s.rt.curSpeed = kmh
}

// SetDirection 设置当前方向
func (s *Simulator) SetDirection(dir uint16) {
	s.rt.mu.Lock()
	defer s.rt.mu.Unlock()
	s.rt.curDir = dir
}

// SetAuthCode 设置运行时鉴权码
func (s *Simulator) SetAuthCode(code string) {
	s.rt.mu.Lock()
	defer s.rt.mu.Unlock()
	s.rt.authCode = code
}

// GetAuthCode 获取运行时鉴权码
func (s *Simulator) GetAuthCode() string {
	s.rt.mu.Lock()
	defer s.rt.mu.Unlock()
	return s.rt.authCode
}

// CurLat 返回当前纬度
func (s *Simulator) CurLat() float64 {
	s.rt.mu.Lock()
	defer s.rt.mu.Unlock()
	return s.rt.curLat
}

// CurLon 返回当前经度
func (s *Simulator) CurLon() float64 {
	s.rt.mu.Lock()
	defer s.rt.mu.Unlock()
	return s.rt.curLon
}

// CurSpeed 返回当前速度
func (s *Simulator) CurSpeed() float64 {
	s.rt.mu.Lock()
	defer s.rt.mu.Unlock()
	return s.rt.curSpeed
}

// CurDir 返回当前方向
func (s *Simulator) CurDir() uint16 {
	s.rt.mu.Lock()
	defer s.rt.mu.Unlock()
	return s.rt.curDir
}

// AdvanceRoute 推进路线并返回新坐标和方向
func (s *Simulator) AdvanceRoute(speedKmh float64) (lat, lon float64, direction uint16) {
	s.rt.mu.Lock()
	s.rt.curSpeed = speedKmh
	direction = s.rt.curDir
	s.rt.mu.Unlock()

	lat, lon = s.MoveAlongRoute()
	return
}
