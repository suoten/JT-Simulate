package base

import (
	"context"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

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

// DeviceConfig 设备配置
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

// MessageHandler 消息处理器
type MessageHandler func(msg *types.Message)

// RawDataHandler 原始数据处理器（子类设置后，recvLoop会把收到的原始数据传给它）
type RawDataHandler func(data []byte)

// Simulator 仿真器基类
type Simulator struct {
	config     *DeviceConfig
	state      atomic.Int32
	conn       net.Conn
	connMu     sync.RWMutex
	seqNum     uint16
	seqMu      sync.Mutex
	cancel     context.CancelFunc
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
	return &Simulator{
		config: cfg,
	}
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
	if s.onState != nil {
		s.onState(state)
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

// NextSeqNum 获取下一个序号
func (s *Simulator) NextSeqNum() uint16 {
	s.seqMu.Lock()
	defer s.seqMu.Unlock()
	s.seqNum++
	return s.seqNum
}

// Connect 连接目标平台
func (s *Simulator) Connect(ctx context.Context) error {
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
	s.wg.Add(1)
	go s.recvLoop(ctx)

	return nil
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
	if s.onSend != nil {
		s.onSend(data)
	}
	s.mu.RUnlock()

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
