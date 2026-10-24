package engine

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/suoten/jt-simulate/internal/simulator/base"
	"github.com/suoten/jt-simulate/pkg/codec/jt808"
	"github.com/suoten/jt-simulate/pkg/types"
)

// trackedPlatform 可控的 JT808 假平台：支持主动踢断全部连接、统计收到的消息
type trackedPlatform struct {
	listener net.Listener
	wg       sync.WaitGroup

	mu       sync.Mutex
	conns    map[net.Conn]struct{}
	totalReg int // 收到的注册(0x0100)总数（单调递增，不随连接关闭回退）
	locQuery int // 收到的位置上报(0x0200)总数
}

func newTrackedPlatform(t *testing.T) *trackedPlatform {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	tp := &trackedPlatform{
		listener: ln,
		conns:    make(map[net.Conn]struct{}),
	}
	tp.wg.Add(1)
	go tp.acceptLoop()
	return tp
}

func (tp *trackedPlatform) addr() string { return tp.listener.Addr().String() }

func (tp *trackedPlatform) acceptLoop() {
	defer tp.wg.Done()
	for {
		conn, err := tp.listener.Accept()
		if err != nil {
			return
		}
		tp.mu.Lock()
		tp.conns[conn] = struct{}{}
		tp.mu.Unlock()
		tp.wg.Add(1)
		go tp.handleConn(conn)
	}
}

func (tp *trackedPlatform) handleConn(conn net.Conn) {
	defer tp.wg.Done()
	defer func() {
		tp.mu.Lock()
		delete(tp.conns, conn)
		tp.mu.Unlock()
		conn.Close()
	}()

	codec := jt808.NewCodec()
	buf := make([]byte, 4096)
	for {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		n, err := conn.Read(buf)
		if err != nil {
			return
		}
		frames := jt808.SplitByDelimiter(buf[:n])
		for _, frame := range frames {
			msg, err := codec.Decode(frame)
			if err != nil {
				continue
			}
			switch msg.Header.MsgID {
			case 0x0100:
				tp.mu.Lock()
				tp.totalReg++
				tp.mu.Unlock()
				respData, _ := codec.Encode(
					&types.MessageHeader{MsgID: 0x8100, Phone: msg.Header.Phone, SeqNum: msg.Header.SeqNum, Version2019: true, ProtocolVer: 1},
					&jt808.RegisterRespMessage{RespSeqNum: msg.Header.SeqNum, Result: 0, AuthCode: "TEST_AUTH"},
				)
				conn.Write(respData)
			case 0x0102:
				respData, _ := codec.Encode(
					&types.MessageHeader{MsgID: 0x8001, Phone: msg.Header.Phone, SeqNum: msg.Header.SeqNum, Version2019: true, ProtocolVer: 1},
					&jt808.PlatformGeneralRespMessage{RespSeqNum: msg.Header.SeqNum, RespMsgID: 0x0102, Result: 0},
				)
				conn.Write(respData)
			case 0x0200:
				tp.mu.Lock()
				tp.locQuery++
				tp.mu.Unlock()
			}
		}
	}
}

// dropAll 踢断当前所有连接（模拟平台故障/网络断开）
func (tp *trackedPlatform) dropAll() {
	tp.mu.Lock()
	conns := make([]net.Conn, 0, len(tp.conns))
	for c := range tp.conns {
		conns = append(conns, c)
	}
	tp.mu.Unlock()
	for _, c := range conns {
		c.Close()
	}
}

// broadcastSplitFrame 把一帧拆成两个 TCP 分片发送（模拟粘包/半包）
func (tp *trackedPlatform) broadcastSplitFrame(data []byte) {
	mid := len(data) / 2
	tp.mu.Lock()
	conns := make([]net.Conn, 0, len(tp.conns))
	for c := range tp.conns {
		conns = append(conns, c)
	}
	tp.mu.Unlock()
	for _, c := range conns {
		c.Write(data[:mid])
		time.Sleep(50 * time.Millisecond)
		c.Write(data[mid:])
	}
}

// connCount 当前存活连接数
func (tp *trackedPlatform) connCount() int {
	tp.mu.Lock()
	defer tp.mu.Unlock()
	return len(tp.conns)
}

// totalRegisters 所有连接收到的注册总数
func (tp *trackedPlatform) totalRegisters() int {
	tp.mu.Lock()
	defer tp.mu.Unlock()
	return tp.totalReg
}

func (tp *trackedPlatform) locationReports() int {
	tp.mu.Lock()
	defer tp.mu.Unlock()
	return tp.locQuery
}

func (tp *trackedPlatform) close() {
	tp.listener.Close()
	tp.dropAll()
	tp.wg.Wait()
}

func waitDeviceState(t *testing.T, e *Engine, id string, want base.DeviceState, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		info, err := e.GetDevice(id)
		if err == nil && info.State() == want {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	info, _ := e.GetDevice(id)
	got := base.DeviceState(-1)
	if info != nil {
		got = info.State()
	}
	t.Fatalf("device %s did not reach %s within %v (now %s)", id, want, timeout, got)
}

func newReconnectTestDevice(addr, id, phone string) *base.DeviceConfig {
	return &base.DeviceConfig{
		ID:                id,
		Protocol:          "jt808",
		Phone:             phone,
		Plate:             "重连测试",
		PlateColor:        1,
		TargetAddr:        addr,
		AuthCode:          "test",
		ProvinceID:        11,
		CityID:            100,
		Manufacturer:      "TEST",
		TerminalModel:     "T-100",
		TerminalID:        "0000001",
		HeartbeatInterval: 60,
		LocationInterval:  1, // 快速位置上报，用于驱动断线检测与重组测试
		ReconnectInterval: 1, // 快速重连
		StartLat:          39.9093,
		StartLon:          116.3974,
	}
}

// TestReconnectAfterPlatformDrop 平台踢断连接后，设备应自动重连并重新注册
// （回归：旧实现重连会取消自身 ctx，第二次连接起设备永久失效/死锁）
func TestReconnectAfterPlatformDrop(t *testing.T) {
	fp := newTrackedPlatform(t)
	defer fp.close()

	e := New()
	e.SetStorage(nil)
	cfg := newReconnectTestDevice(fp.addr(), "rc-1", "013800006666")
	if err := e.CreateDevice(cfg); err != nil {
		t.Fatalf("CreateDevice failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := e.StartDevice(ctx, "rc-1"); err != nil {
		t.Fatalf("StartDevice failed: %v", err)
	}
	waitDeviceState(t, e, "rc-1", base.StateOnline, 10*time.Second)

	regBefore := fp.totalRegisters()

	// 平台踢断全部连接
	fp.dropAll()
	waitDeviceState(t, e, "rc-1", base.StateOffline, 10*time.Second)

	// 设备应在 ReconnectInterval 内自动重连并重新注册
	waitDeviceState(t, e, "rc-1", base.StateOnline, 15*time.Second)

	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if fp.totalRegisters() > regBefore {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if got := fp.totalRegisters(); got <= regBefore {
		t.Errorf("expected re-registration after reconnect (registers %d -> >%d)", regBefore, got)
	}

	stats := e.GetStats()
	if stats.OnlineDevices != 1 {
		t.Errorf("expected OnlineDevices=1 after reconnect, got %d", stats.OnlineDevices)
	}

	e.StopAll()
	waitDeviceState(t, e, "rc-1", base.StateOffline, 5*time.Second)
	if stats := e.GetStats(); stats.OnlineDevices != 0 {
		t.Errorf("expected OnlineDevices=0 after StopAll, got %d", stats.OnlineDevices)
	}
}

// TestStickyFrameReassembly 跨 TCP 分片的下行帧应被重组后处理：
// 平台把一帧拆成两半发送，设备仍应正常应答（回归：旧实现直接丢帧）
func TestStickyFrameReassembly(t *testing.T) {
	fp := newTrackedPlatform(t)
	defer fp.close()

	e := New()
	e.SetStorage(nil)
	cfg := newReconnectTestDevice(fp.addr(), "rc-2", "013800005555")
	if err := e.CreateDevice(cfg); err != nil {
		t.Fatalf("CreateDevice failed: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := e.StartDevice(ctx, "rc-2"); err != nil {
		t.Fatalf("StartDevice failed: %v", err)
	}
	waitDeviceState(t, e, "rc-2", base.StateOnline, 10*time.Second)

	// 收到若干正常位置上报（位置上报每 1 秒一次，作为基线）
	time.Sleep(3 * time.Second)
	baseCount := fp.locationReports()

	// 平台分两半下发 0x8201 位置查询——设备应重组后回复 0x0200
	codec := jt808.NewCodec()
	query, err := codec.Encode(
		&types.MessageHeader{MsgID: 0x8201, Phone: cfg.Phone, SeqNum: 1, Version2019: true, ProtocolVer: 1},
		&jt808.LocationQueryMessage{},
	)
	if err != nil {
		t.Fatalf("encode query failed: %v", err)
	}
	fp.broadcastSplitFrame(query)

	// 位置查询应答会推高 0x0200 计数（相对 1 秒周期上报的基线，容差放宽）
	deadline := time.Now().Add(10 * time.Second)
	ok := false
	for time.Now().Before(deadline) {
		if fp.locationReports() > baseCount {
			ok = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if !ok {
		t.Errorf("split frame was not reassembled: location reports stayed at %d", baseCount)
	}

	e.StopAll()
}

// TestDoubleStartProtection 并发双启动应被拒绝，且不会产生双连接
func TestDoubleStartProtection(t *testing.T) {
	fp := newTrackedPlatform(t)
	defer fp.close()

	e := New()
	e.SetStorage(nil)
	cfg := newReconnectTestDevice(fp.addr(), "rc-3", "013800004444")
	if err := e.CreateDevice(cfg); err != nil {
		t.Fatalf("CreateDevice failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	results := make([]error, 4)
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx] = e.StartDevice(ctx, "rc-3")
		}(i)
	}
	wg.Wait()

	success := 0
	for _, err := range results {
		if err == nil {
			success++
		}
	}
	if success != 1 {
		t.Errorf("expected exactly 1 successful concurrent start, got %d (errs: %v)", success, results)
	}

	waitDeviceState(t, e, "rc-3", base.StateOnline, 10*time.Second)

	// 单会话：平台侧只应有一条连接
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) && fp.connCount() != 1 {
		time.Sleep(100 * time.Millisecond)
	}
	if got := fp.connCount(); got != 1 {
		t.Errorf("expected exactly 1 platform-side connection, got %d", got)
	}

	// 在线计数不应因多次启动尝试而漂移
	if stats := e.GetStats(); stats.OnlineDevices != 1 {
		t.Errorf("expected OnlineDevices=1, got %d", stats.OnlineDevices)
	}

	e.StopAll()
	waitDeviceState(t, e, "rc-3", base.StateOffline, 5*time.Second)
}

// TestStartAfterFailedStart 拨号失败后应能再次启动（状态回滚正确）
func TestStartAfterFailedStart(t *testing.T) {
	e := New()
	e.SetStorage(nil)
	cfg := newReconnectTestDevice("127.0.0.1:1", "rc-4", "013800003333")
	cfg.ReconnectInterval = 3600 // 失败后不要快速重试干扰测试
	if err := e.CreateDevice(cfg); err != nil {
		t.Fatalf("CreateDevice failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := e.StartDevice(ctx, "rc-4"); err == nil {
		t.Fatal("expected dial failure error, got nil")
	}
	waitDeviceState(t, e, "rc-4", base.StateOffline, 5*time.Second)

	// 修正目标地址后再次启动应成功
	fp := newTrackedPlatform(t)
	defer fp.close()
	info, _ := e.GetDevice("rc-4")
	info.Config.TargetAddr = fp.addr()
	if err := e.StartDevice(ctx, "rc-4"); err != nil {
		t.Fatalf("restart after failure failed: %v", err)
	}
	waitDeviceState(t, e, "rc-4", base.StateOnline, 10*time.Second)

	stats := e.GetStats()
	if stats.OnlineDevices != 1 {
		t.Errorf("expected OnlineDevices=1, got %d", stats.OnlineDevices)
	}

	e.StopAll()
}
