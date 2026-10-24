package base

import (
	"context"
	"errors"
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

	Route    string  `json:"route"`
	StartLat float64 `json:"start_lat"`
	StartLon float64 `json:"start_lon"`
}

// runtimeState 运行时状态（可变，需要加锁保护）
type runtimeState struct {
	mu       sync.Mutex
	curLat   float64
	curLon   float64
	curSpeed float64 // km/h
	curDir   uint16  // 0-359
	authCode string  // 运行时从平台收到的鉴权码
}

// MessageHandler 消息处理器
type MessageHandler func(msg *types.Message)

// RawDataHandler 原始数据处理器（子类设置后，recvLoop会把收到的原始数据传给它）
type RawDataHandler func(data []byte)

// FrameSplitter 帧分割器：从累计接收缓冲中提取完整帧。
// 返回 frames 为完整帧列表（可直接交给 RawDataHandler/解码器），
// restLen 为缓冲末尾属于残缺帧、需要保留到下次数据到达的字节数。
type FrameSplitter func(buf []byte) (frames [][]byte, restLen int)

const (
	// dialTimeout 拨号超时：防止目标不可达时调用方被 OS 级 TCP 超时（可达 2 分钟）阻塞
	dialTimeout = 15 * time.Second
	// writeTimeout 单次发送超时：平台停止读取导致发送缓冲堆满时，避免发送方永久阻塞
	writeTimeout = 30 * time.Second
	// readIdleTimeout 空闲读超时（安全网）。链路活性主要依赖 TCP KeepAlive 与写超时，
	// 平台长时间无下行（纯监听型平台）不应被误判为断线，因此取值远大于心跳间隔。
	readIdleTimeout = 10 * time.Minute
	// maxRecvBuffer 帧重组缓冲上限：超过视为脏数据流，丢弃以免内存无限增长
	maxRecvBuffer = 256 * 1024
)

// Simulator 仿真器基类
type Simulator struct {
	config *DeviceConfig
	rt     runtimeState // 运行时状态（带锁）
	state  atomic.Int32
	conn   net.Conn
	connMu sync.RWMutex
	seqNum uint16
	seqMu  sync.Mutex

	cancelMu   sync.Mutex
	sessionActive bool             // 是否存在进行中的会话（Connect 置位，Disconnect 清除）
	rootCtx    context.Context    // 会话生命周期：Connect 创建，Disconnect 取消
	rootCancel context.CancelFunc
	connCtx    context.Context    // 当前连接生命周期：connectAndRun 创建，断线/被替换时取消
	connCancel context.CancelFunc

	wg            sync.WaitGroup
	onRecv        MessageHandler
	onRawRecv     RawDataHandler
	frameSplitter FrameSplitter
	onState       func(from, to DeviceState)
	onSend        func([]byte)
	mu            sync.RWMutex
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

// SetState 设置状态（状态实际变化时才触发回调，回调携带旧→新状态，
// 上层可据此做边沿触发式统计，避免同一事件重复计数）
func (s *Simulator) SetState(newState DeviceState) {
	for {
		old := DeviceState(s.state.Load())
		if old == newState {
			return
		}
		if s.state.CompareAndSwap(int32(old), int32(newState)) {
			s.fireStateChange(old, newState)
			return
		}
	}
}

// TrySetState 原子状态迁移：仅当当前状态为 from 时迁移到 to。
// 返回是否成功。用于并发场景下的单次启动保护（check-then-act 竞态）。
func (s *Simulator) TrySetState(from, to DeviceState) bool {
	if !s.state.CompareAndSwap(int32(from), int32(to)) {
		return false
	}
	s.fireStateChange(from, to)
	return true
}

func (s *Simulator) fireStateChange(from, to DeviceState) {
	s.mu.RLock()
	fn := s.onState
	s.mu.RUnlock()
	if fn != nil {
		fn(from, to)
	}
}

// SetOnRecv 设置接收回调
func (s *Simulator) SetOnRecv(handler MessageHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onRecv = handler
}

// SetOnState 设置状态变更回调
func (s *Simulator) SetOnState(handler func(from, to DeviceState)) {
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

// SetFrameSplitter 设置帧分割器（启用 TCP 粘包/半包重组：
// recvLoop 会累计数据并只把完整帧交给 onRawRecv）。未设置时按原始数据块直通。
func (s *Simulator) SetFrameSplitter(fn FrameSplitter) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.frameSplitter = fn
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

// Connect 建立到目标平台的连接并启动接收循环。
// ctx 作为本会话的父 context：其取消会级联停止本设备的全部 goroutine。
// 若上一次会话仍在运行，会先将其彻底断开（幂等），保证同一设备同一时刻只有一条会话。
func (s *Simulator) Connect(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	s.Disconnect()

	s.cancelMu.Lock()
	s.sessionActive = true
	s.rootCtx, s.rootCancel = context.WithCancel(ctx)
	rootCtx := s.rootCtx
	s.cancelMu.Unlock()

	return s.connectAndRun(rootCtx)
}

// connectAndRun 拨号并启动接收循环
func (s *Simulator) connectAndRun(rootCtx context.Context) error {
	s.SetState(StateConnecting)

	d := net.Dialer{
		Timeout:   dialTimeout,
		KeepAlive: 30 * time.Second, // TCP 保活，检测半开连接
	}
	conn, err := d.DialContext(rootCtx, "tcp", s.config.TargetAddr)
	if err != nil {
		s.SetState(StateOffline)
		return fmt.Errorf("connect %s: %w", s.config.TargetAddr, err)
	}

	// 替换旧连接（重连场景）并关闭，避免句柄泄漏
	s.connMu.Lock()
	old := s.conn
	s.conn = conn
	s.connMu.Unlock()
	if old != nil && old != conn {
		old.Close()
	}

	// 为本连接派生独立 ctx：连接断开或被新连接替换时取消，
	// 联动停止挂在该连接上的心跳/位置上报 goroutine
	connCtx, connCancel := context.WithCancel(rootCtx)
	s.cancelMu.Lock()
	prev := s.connCancel
	s.connCtx = connCtx
	s.connCancel = connCancel
	s.cancelMu.Unlock()
	if prev != nil {
		prev()
	}

	s.wg.Add(1)
	go s.recvLoop(connCtx, conn)

	s.SetState(StateOnline)
	return nil
}

// Context 返回当前连接的 context（连接断开时被取消）。
// 子类应在每次连接建立后（含重连回调内）重新用它启动心跳/位置上报等 goroutine。
func (s *Simulator) Context() context.Context {
	s.cancelMu.Lock()
	defer s.cancelMu.Unlock()
	if s.connCtx != nil {
		return s.connCtx
	}
	if s.rootCtx != nil {
		return s.rootCtx
	}
	return context.Background()
}

// StartReconnectLoop 启动自动重连循环。
// 循环挂在本会话的 root ctx 上（跨连接存活），onReconnect 收到的 ctx
// 是重连成功后"当前连接"的 ctx，再次断线时会被取消，回调内的等待应及时响应。
func (s *Simulator) StartReconnectLoop(onReconnect func(ctx context.Context) error) {
	s.cancelMu.Lock()
	rootCtx := s.rootCtx
	s.cancelMu.Unlock()
	if rootCtx == nil {
		return
	}
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for {
			if rootCtx.Err() != nil {
				return
			}
			if s.State() != StateOffline {
				select {
				case <-rootCtx.Done():
					return
				case <-time.After(5 * time.Second):
				}
				continue
			}

			interval := s.config.ReconnectInterval
			if interval <= 0 {
				interval = 15
			}
			select {
			case <-rootCtx.Done():
				return
			case <-time.After(time.Duration(interval) * time.Second):
			}

			if s.State() != StateOffline || rootCtx.Err() != nil {
				continue
			}
			logger.Info("尝试重连", "phone", s.config.Phone, "target", s.config.TargetAddr)
			if err := s.connectAndRun(rootCtx); err != nil {
				logger.Error("重连失败", "phone", s.config.Phone, "error", err)
				continue
			}
			if onReconnect != nil {
				if err := onReconnect(s.Context()); err != nil {
					logger.Error("重连后重新上线失败，拆除连接等待下次重连", "phone", s.config.Phone, "error", err)
					// 不能调用 Disconnect：其内部 wg.Wait 会等待本 goroutine 自身，死锁。
					// 只拆除当前连接，保持重连循环存活。
					s.teardownConnection()
					continue
				}
			}
			logger.Info("重连成功", "phone", s.config.Phone)
		}
	}()
}

// teardownConnection 关闭当前连接并取消其 ctx（不等待 goroutine 退出）。
// 供 recvLoop 退出路径与重连循环内部使用——这两处运行在 wg 成员 goroutine 中，
// 禁止调用会 wg.Wait 的 Disconnect。
func (s *Simulator) teardownConnection() {
	s.cancelMu.Lock()
	cancel := s.connCancel
	s.connCancel = nil
	s.cancelMu.Unlock()
	if cancel != nil {
		cancel()
	}

	s.connMu.Lock()
	conn := s.conn
	s.conn = nil
	s.connMu.Unlock()
	if conn != nil {
		conn.Close()
	}
	s.SetState(StateOffline)
}

// Disconnect 彻底断开：取消会话 root ctx（联动停止重连/心跳/位置/接收 goroutine），
// 关闭连接并等待所有 goroutine 退出。对空闲设备调用是无副作用的 no-op。
// 注意：禁止在设备的回调或 goroutine 内部调用（wg.Wait 会等待调用者自身导致死锁），
// 这类场景请使用 teardownConnection。
func (s *Simulator) Disconnect() {
	s.cancelMu.Lock()
	if !s.sessionActive {
		s.cancelMu.Unlock()
		return
	}
	s.sessionActive = false
	rootCancel := s.rootCancel
	connCancel := s.connCancel
	s.cancelMu.Unlock()

	s.SetState(StateClosing)
	// 必须在 wg.Wait 之前取消两个 ctx：重连循环/心跳/接收等 wg 成员
	// 依赖 ctx.Done 退出，先等后取消会形成循环等待（死锁）
	if connCancel != nil {
		connCancel()
	}
	if rootCancel != nil {
		rootCancel()
	}
	s.connMu.Lock()
	conn := s.conn
	s.conn = nil
	s.connMu.Unlock()
	if conn != nil {
		conn.Close() // 强制阻塞中的 Read/Write 立即返回
	}

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

	conn.SetWriteDeadline(time.Now().Add(writeTimeout))
	if _, err := conn.Write(data); err != nil {
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

// recvLoop 接收循环（含 TCP 粘包/半包重组）。
// conn 是本循环独占的连接：退出时负责关闭它，保证任何路径都不泄漏句柄。
func (s *Simulator) recvLoop(ctx context.Context, conn net.Conn) {
	defer s.wg.Done()
	defer conn.Close()

	buf := make([]byte, 8192)
	var accum []byte
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		conn.SetReadDeadline(time.Now().Add(readIdleTimeout))
		n, err := conn.Read(buf)
		if n > 0 {
			accum = append(accum, buf[:n]...)
			s.deliverFrames(&accum)
		}
		if err != nil {
			if ctx.Err() != nil || errors.Is(err, net.ErrClosed) {
				// 主动断开（Disconnect/连接被新连接替换）：状态由调用方设置
				return
			}
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				logger.Warn("连接长时间无下行数据，视为断线", "phone", s.config.Phone, "idle", readIdleTimeout.String())
			} else {
				logger.Debug("连接断开", "phone", s.config.Phone, "error", err)
			}
			// 连接异常断开：拆除连接（触发重连循环），本接收循环结束。
			// 仅当自己仍是当前连接时才拆（若已被重连替换，拆除的应是旧连接自身，
			// 不应误杀新连接）
			s.connMu.RLock()
			cur := s.conn
			s.connMu.RUnlock()
			if cur == conn {
				s.teardownConnection()
			}
			return
		}
	}
}

// deliverFrames 用帧分割器从累计缓冲中提取完整帧交给 onRawRecv，
// 并保留残缺尾部等待后续数据。未设置分割器时按原始数据块直通（无分隔符协议）。
func (s *Simulator) deliverFrames(accum *[]byte) {
	s.mu.RLock()
	splitter := s.frameSplitter
	handler := s.onRawRecv
	s.mu.RUnlock()

	if handler == nil {
		*accum = (*accum)[:0]
		return
	}

	if splitter == nil {
		data := make([]byte, len(*accum))
		copy(data, *accum)
		*accum = (*accum)[:0]
		handler(data)
		return
	}

	frames, keep := splitter(*accum)
	if keep > len(*accum) {
		keep = len(*accum)
	}
	if keep > 0 {
		rest := make([]byte, keep)
		copy(rest, (*accum)[len(*accum)-keep:])
		*accum = rest
	} else {
		*accum = (*accum)[:0]
	}
	if len(*accum) > maxRecvBuffer {
		logger.Warn("接收重组缓冲溢出，丢弃未成帧数据", "phone", s.config.Phone, "size", len(*accum))
		*accum = (*accum)[:0]
	}
	for _, frame := range frames {
		handler(frame)
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
