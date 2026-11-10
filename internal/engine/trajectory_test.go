package engine

import (
	"context"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/suoten/jt-simulate/internal/simulator/base"
	jt808sim "github.com/suoten/jt-simulate/internal/simulator/jt808"
	"github.com/suoten/jt-simulate/internal/trajectory"
	"github.com/suoten/jt-simulate/pkg/codec/jt808"
	"github.com/suoten/jt-simulate/pkg/types"
)

// locRecorder 模拟 JT808 平台并记录收到的 0x0200 位置上报
type locRecorder struct {
	listener net.Listener
	wg       sync.WaitGroup

	mu    sync.Mutex
	locs  []*jt808.LocationMessage
	conns []net.Conn
}

func newLocRecorder(t *testing.T) *locRecorder {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	lr := &locRecorder{listener: ln}
	lr.wg.Add(1)
	go lr.acceptLoop()
	return lr
}

func (lr *locRecorder) addr() string { return lr.listener.Addr().String() }

func (lr *locRecorder) locations() []*jt808.LocationMessage {
	lr.mu.Lock()
	defer lr.mu.Unlock()
	out := make([]*jt808.LocationMessage, len(lr.locs))
	copy(out, lr.locs)
	return out
}

func (lr *locRecorder) acceptLoop() {
	defer lr.wg.Done()
	for {
		conn, err := lr.listener.Accept()
		if err != nil {
			return
		}
		lr.mu.Lock()
		lr.conns = append(lr.conns, conn)
		lr.mu.Unlock()
		lr.wg.Add(1)
		go lr.handleConn(conn)
	}
}

func (lr *locRecorder) handleConn(conn net.Conn) {
	defer lr.wg.Done()
	defer conn.Close()

	codec := jt808.NewCodec()
	buf := make([]byte, 4096)

	for {
		conn.SetReadDeadline(time.Now().Add(30 * time.Second))
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
				respData, _ := codec.Encode(
					&types.MessageHeader{
						MsgID:       0x8100,
						Phone:       msg.Header.Phone,
						SeqNum:      msg.Header.SeqNum,
						Version2019: true,
						ProtocolVer: 1,
					},
					&jt808.RegisterRespMessage{
						RespSeqNum: msg.Header.SeqNum,
						Result:     0,
						AuthCode:   "TEST_AUTH_CODE",
					},
				)
				conn.Write(respData)
			case 0x0102:
				respData, _ := codec.Encode(
					&types.MessageHeader{
						MsgID:       0x8001,
						Phone:       msg.Header.Phone,
						SeqNum:      msg.Header.SeqNum,
						Version2019: true,
						ProtocolVer: 1,
					},
					&jt808.PlatformGeneralRespMessage{
						RespSeqNum: msg.Header.SeqNum,
						RespMsgID:  0x0102,
						Result:     0,
					},
				)
				conn.Write(respData)
			case 0x0200:
				if loc, ok := msg.Body.(*jt808.LocationMessage); ok {
					lr.mu.Lock()
					lr.locs = append(lr.locs, loc)
					lr.mu.Unlock()
				}
			}
		}
	}
}

func (lr *locRecorder) close() {
	lr.listener.Close()
	lr.mu.Lock()
	for _, c := range lr.conns {
		c.Close()
	}
	lr.mu.Unlock()
	lr.wg.Wait()
}

// waitForCond 轮询直到条件满足或超时
func waitForCond(desc string, timeout time.Duration, cond func() bool) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return nil
		}
		time.Sleep(200 * time.Millisecond)
	}
	return fmt.Errorf("等待超时: %s", desc)
}

func newTrajTestDevice(t *testing.T, e *Engine, lr *locRecorder, id, protocol string) {
	cfg := &base.DeviceConfig{
		ID:                id,
		Protocol:          protocol,
		Phone:             "01380000" + id[len(id)-4:],
		Plate:             "轨迹测试",
		PlateColor:        1,
		TargetAddr:        lr.addr(),
		AuthCode:          "test_auth",
		ProvinceID:        11,
		CityID:            100,
		Manufacturer:      "TEST",
		TerminalModel:     "T-100",
		TerminalID:        "0000001",
		HeartbeatInterval: 60,
		LocationInterval:  1, // 默认上报调快，用于验证回放期间被抑制
		ReconnectInterval: 15,
		StartLat:          30.0,
		StartLon:          120.0,
	}
	if err := e.CreateDevice(cfg); err != nil {
		t.Fatalf("CreateDevice failed: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := e.StartDevice(ctx, id); err != nil {
		t.Fatalf("StartDevice failed: %v", err)
	}
	info, err := e.GetDevice(id)
	if err != nil {
		t.Fatalf("GetDevice failed: %v", err)
	}
	if info.State() != base.StateOnline {
		t.Fatalf("device should be online, got %s", info.State())
	}
}

// straightTimedRoute 南北向直线：39.90 → 39.91 → 39.92（约 2220 米，带速度标记）
func straightTrajRoute(t *testing.T) *trajectory.Route {
	r, err := trajectory.NewRoute("测试直线", "text", []trajectory.Point{
		{Lat: 39.90, Lon: 116.4, SpeedKmh: 36},
		{Lat: 39.91, Lon: 116.4, SpeedKmh: 36},
		{Lat: 39.92, Lon: 116.4},
	})
	if err != nil {
		t.Fatalf("NewRoute failed: %v", err)
	}
	return r
}

// TestTrajectoryPlaybackUniform 匀速推进模式端到端：
// 1) 回放期间默认周期上报被抑制，0x0200 只来自回放引擎且经纬度沿轨迹推进
// 2) 非循环回放结束后默认上报从轨迹终点恢复
func TestTrajectoryPlaybackUniform(t *testing.T) {
	lr := newLocRecorder(t)
	defer lr.close()

	e := New()
	e.SetStorage(nil)
	newTrajTestDevice(t, e, lr, "traj-0001", "jt808")

	te := NewTrajectoryEngine(e.GetDevice)
	defer te.StopAll()

	route := straightTrajRoute(t)
	// 36km/h = 10m/s，间隔 1s → 每秒推进约 10 米；全程约 222 秒
	if err := te.Start(route, TrajOptions{
		DeviceID:    "traj-0001",
		RouteName:   route.Name,
		IntervalSec: 1,
		Mode:        TrajModeUniform,
		SpeedKmh:    36,
	}); err != nil {
		t.Fatalf("Trajectory Start failed: %v", err)
	}

	// 回放中：等 3.5 秒，应收到 ≥4 条 0x0200（起点 + 3 个 tick）
	if err := waitForCond("回放产生至少4条0x0200", 6*time.Second, func() bool {
		return len(lr.locations()) >= 4
	}); err != nil {
		t.Fatal(err)
	}

	locs := lr.locations()
	// 经纬度应沿轨迹向北推进（起点 39.90）
	last := locs[len(locs)-1]
	if last.Latitude <= 39.90 {
		t.Errorf("回放位置应从起点向北推进，got lat=%f", last.Latitude)
	}
	// 方向应约为正北 0°（环绕判断：仅当落在 2-358 之间才算偏离）
	if last.Direction > 2 && last.Direction < 358 {
		t.Errorf("向北行驶方向应≈0，got %d", last.Direction)
	}
	// 速度应为轨迹文件自带 36km/h（0.1km/h 单位 360）
	if last.Speed < 355 || last.Speed > 365 {
		t.Errorf("速度应≈360(0.1km/h)，got %d", last.Speed)
	}

	// 回放会话存在且 UsesRoute 命中
	if te.RunningCount() != 1 {
		t.Errorf("RunningCount 应为 1，got %d", te.RunningCount())
	}
	if !te.UsesRoute(route.Name) {
		t.Error("UsesRoute 应命中回放中的轨迹")
	}

	// 停止回放
	if err := te.Stop("traj-0001"); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}
	if te.RunningCount() != 0 {
		t.Errorf("停止后 RunningCount 应为 0，got %d", te.RunningCount())
	}

	// 停止后默认周期上报应恢复（LocationInterval=1s）：再收到若干条 0x0200
	countAtStop := len(lr.locations())
	if err := waitForCond("停止后默认上报恢复", 5*time.Second, func() bool {
		return len(lr.locations()) > countAtStop
	}); err != nil {
		t.Fatal(err)
	}
}

// TestTrajectoryPlaybackTimelineLoop 按时间回放 + 循环：时钟回绕应产生圈数
func TestTrajectoryPlaybackTimelineLoop(t *testing.T) {
	lr := newLocRecorder(t)
	defer lr.close()

	e := New()
	e.SetStorage(nil)
	newTrajTestDevice(t, e, lr, "traj-0002", "jt808")

	te := NewTrajectoryEngine(e.GetDevice)
	defer te.StopAll()

	// 4 秒时间轴的环线（首尾相接），间隔 1s、倍速 1 → 每 4 秒回绕一圈
	r, err := trajectory.NewRoute("测试环线", "text", []trajectory.Point{
		{Lat: 39.90, Lon: 116.4, TimeSec: 0},
		{Lat: 39.91, Lon: 116.4, TimeSec: 4},
		{Lat: 39.90, Lon: 116.4, TimeSec: 8},
	})
	if err != nil {
		t.Fatalf("NewRoute failed: %v", err)
	}
	if !r.HasTime {
		t.Fatal("环线应有时间轴")
	}

	if err := te.Start(r, TrajOptions{
		DeviceID:    "traj-0002",
		RouteName:   r.Name,
		IntervalSec: 1,
		Mode:        TrajModeTimeline,
		SpeedFactor: 1,
		Loop:        true,
	}); err != nil {
		t.Fatalf("Trajectory Start failed: %v", err)
	}

	// 9 秒后应至少完成 1 圈（时钟在 4s、8s 回绕）
	if err := waitForCond("循环回放产生至少1圈", 11*time.Second, func() bool {
		return te.Status()[0].LoopsDone >= 1
	}); err != nil {
		t.Fatal(err)
	}

	st := te.Status()
	if len(st) != 1 {
		t.Fatalf("Status 应含 1 个会话，got %d", len(st))
	}
	if st[0].PointsSent < 8 {
		t.Errorf("循环回放累计上报应≥8 条，got %d", st[0].PointsSent)
	}

	// 位置应随时间轴往返（收到过 >39.905 与 <39.905 的点）
	lo, hi := false, false
	for _, loc := range lr.locations() {
		if loc.Latitude < 39.905 {
			lo = true
		}
		if loc.Latitude > 39.905 {
			hi = true
		}
	}
	if !lo || !hi {
		t.Errorf("时间回放位置应往返变化，lo=%v hi=%v", lo, hi)
	}
}

// TestTrajectoryStartValidation 启动校验：离线设备 / 非法模式 / 非 JT808 协议
func TestTrajectoryStartValidation(t *testing.T) {
	lr := newLocRecorder(t)
	defer lr.close()

	e := New()
	e.SetStorage(nil)
	te := NewTrajectoryEngine(e.GetDevice)
	defer te.StopAll()

	route := straightTrajRoute(t)

	// 设备不存在
	if err := te.Start(route, TrajOptions{DeviceID: "no-such", RouteName: route.Name}); err == nil {
		t.Error("不存在的设备应报错")
	}

	// 创建但不启动（离线）
	cfg := &base.DeviceConfig{
		ID:         "traj-off",
		Protocol:   "jt808",
		Phone:      "013800009999",
		TargetAddr: lr.addr(),
	}
	if err := e.CreateDevice(cfg); err != nil {
		t.Fatalf("CreateDevice failed: %v", err)
	}
	if err := te.Start(route, TrajOptions{DeviceID: "traj-off", RouteName: route.Name}); err == nil {
		t.Error("离线设备应报错")
	}

	// 非 JT808 系协议
	cfgGbt := &base.DeviceConfig{
		ID:         "traj-gbt",
		Protocol:   "gbt32960",
		Phone:      "VINTEST00001",
		TargetAddr: lr.addr(),
	}
	if err := e.CreateDevice(cfgGbt); err != nil {
		t.Fatalf("CreateDevice gbt failed: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	_ = e.StartDevice(ctx, "traj-gbt") // gbt32960 上线与否不影响校验结果
	if err := te.Start(route, TrajOptions{DeviceID: "traj-gbt", RouteName: route.Name}); err == nil {
		t.Error("gbt32960 设备应报错")
	}

	// 时间模式用于无时间轨迹
	newTrajTestDevice(t, e, lr, "traj-0003", "jt808")
	if err := te.Start(route, TrajOptions{DeviceID: "traj-0003", RouteName: route.Name, Mode: TrajModeTimeline}); err == nil {
		t.Error("无时间轨迹用时间模式应报错")
	}

	// 未知模式
	if err := te.Start(route, TrajOptions{DeviceID: "traj-0003", RouteName: route.Name, Mode: "bogus"}); err == nil {
		t.Error("未知模式应报错")
	}

	if te.RunningCount() != 0 {
		t.Errorf("校验失败场景不应产生会话，got %d", te.RunningCount())
	}
}

// TestTrajectoryDeviceDeleteEndsSession 删除设备后回放会话自动结束
func TestTrajectoryDeviceDeleteEndsSession(t *testing.T) {
	lr := newLocRecorder(t)
	defer lr.close()

	e := New()
	e.SetStorage(nil)
	newTrajTestDevice(t, e, lr, "traj-0004", "jt808")

	te := NewTrajectoryEngine(e.GetDevice)
	defer te.StopAll()

	route := straightTrajRoute(t)
	if err := te.Start(route, TrajOptions{
		DeviceID:    "traj-0004",
		RouteName:   route.Name,
		IntervalSec: 1,
		Mode:        TrajModeUniform,
		SpeedKmh:    36,
	}); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// 删除设备 → 会话应自行结束
	if err := e.DeleteDevice("traj-0004"); err != nil {
		t.Fatalf("DeleteDevice failed: %v", err)
	}
	if err := waitForCond("删除设备后会话结束", 5*time.Second, func() bool {
		return te.RunningCount() == 0
	}); err != nil {
		t.Fatal(err)
	}

	// 同一设备 ID 重新创建并在线后可再次回放（Stop 幂等）
	e.CreateDevice(&base.DeviceConfig{
		ID:                "traj-0004",
		Protocol:          "jt808",
		Phone:             "01380000004",
		TargetAddr:        lr.addr(),
		HeartbeatInterval: 60,
		LocationInterval:  1,
		ReconnectInterval: 15,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := e.StartDevice(ctx, "traj-0004"); err != nil {
		t.Fatalf("重新 StartDevice failed: %v", err)
	}
	if err := te.Start(route, TrajOptions{
		DeviceID:    "traj-0004",
		RouteName:   route.Name,
		IntervalSec: 1,
		Mode:        TrajModeUniform,
		SpeedKmh:    36,
	}); err != nil {
		t.Fatalf("重新 Start failed: %v", err)
	}
	if err := te.Stop("traj-0004"); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}
}

// 编译期断言 jt808sim.Simulator 实现了 Simulator 接口（轨迹引擎的类型断言前提）
var _ Simulator = (*jt808sim.Simulator)(nil)
