package engine

import (
	"context"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/suoten/jt-simulate/internal/simulator/base"
	"github.com/suoten/jt-simulate/pkg/codec/jt808"
	"github.com/suoten/jt-simulate/pkg/types"
)

// fakePlatform 模拟一个 JT808 平台，接收终端连接并回复注册/鉴权应答
type fakePlatform struct {
	listener net.Listener
	wg       sync.WaitGroup
}

func newFakePlatform(t *testing.T) *fakePlatform {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	fp := &fakePlatform{listener: ln}
	fp.wg.Add(1)
	go fp.acceptLoop()
	return fp
}

func (fp *fakePlatform) addr() string {
	return fp.listener.Addr().String()
}

func (fp *fakePlatform) acceptLoop() {
	defer fp.wg.Done()
	for {
		conn, err := fp.listener.Accept()
		if err != nil {
			return
		}
		fp.wg.Add(1)
		go fp.handleConn(conn)
	}
}

func (fp *fakePlatform) handleConn(conn net.Conn) {
	defer fp.wg.Done()
	defer conn.Close()

	codec := jt808.NewCodec()
	buf := make([]byte, 4096)

	for {
		conn.SetReadDeadline(time.Now().Add(10 * time.Second))
		n, err := conn.Read(buf)
		if err != nil {
			return
		}
		if n == 0 {
			continue
		}

		frames := jt808.SplitByDelimiter(buf[:n])
		for _, frame := range frames {
			msg, err := codec.Decode(frame)
			if err != nil {
				continue
			}

			switch msg.Header.MsgID {
			case 0x0100: // 注册——回复注册应答 0x8100
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

			case 0x0102: // 鉴权——回复通用应答 0x8001
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
			}
		}
	}
}

func (fp *fakePlatform) close() {
	fp.listener.Close()
	fp.wg.Wait()
}

// TestStartStopDeviceWithFakePlatform 测试完整的创建→启动→停止链路
func TestStartStopDeviceWithFakePlatform(t *testing.T) {
	fp := newFakePlatform(t)
	defer fp.close()

	e := New()
	e.SetStorage(nil)

	cfg := &base.DeviceConfig{
		ID:                "it-test-001",
		Protocol:          "jt808",
		Phone:             "013800008888",
		Plate:             "集成测试",
		PlateColor:        1,
		TargetAddr:        fp.addr(),
		AuthCode:          "test_auth",
		ProvinceID:        11,
		CityID:            100,
		Manufacturer:      "TEST",
		TerminalModel:     "T-100",
		TerminalID:        "0000001",
		HeartbeatInterval: 60,
		LocationInterval:  30,
		ReconnectInterval: 15,
		StartLat:          39.9093,
		StartLon:          116.3974,
	}

	if err := e.CreateDevice(cfg); err != nil {
		t.Fatalf("CreateDevice failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := e.StartDevice(ctx, "it-test-001"); err != nil {
		t.Fatalf("StartDevice failed: %v", err)
	}

	// 等待设备上线（注册+鉴权）
	time.Sleep(3 * time.Second)

	info, err := e.GetDevice("it-test-001")
	if err != nil {
		t.Fatalf("GetDevice failed: %v", err)
	}
	if info.State() != base.StateOnline {
		t.Errorf("expected StateOnline, got %s", info.State())
	}

	stats := e.GetStats()
	if stats.OnlineDevices != 1 {
		t.Errorf("expected OnlineDevices=1, got %d", stats.OnlineDevices)
	}
	if stats.TotalMessages == 0 {
		t.Error("expected TotalMessages > 0 after registration")
	}

	if err := e.StopDevice("it-test-001"); err != nil {
		t.Fatalf("StopDevice failed: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	info, _ = e.GetDevice("it-test-001")
	if info.State() != base.StateOffline {
		t.Errorf("expected StateOffline after stop, got %s", info.State())
	}

	stats = e.GetStats()
	if stats.OnlineDevices != 0 {
		t.Errorf("expected OnlineDevices=0 after stop, got %d", stats.OnlineDevices)
	}
}

// TestStopAllWithRunningDevices 测试 StopAll 停止多个运行中的设备
func TestStopAllWithRunningDevices(t *testing.T) {
	fp := newFakePlatform(t)
	defer fp.close()

	e := New()
	e.SetStorage(nil)

	for i := 0; i < 3; i++ {
		cfg := &base.DeviceConfig{
			ID:                fmt.Sprintf("stopall-%d", i),
			Protocol:          "jt808",
			Phone:             fmt.Sprintf("01380000%04d", i),
			Plate:             fmt.Sprintf("测%d", i),
			PlateColor:        1,
			TargetAddr:        fp.addr(),
			AuthCode:          "test",
			HeartbeatInterval: 60,
			LocationInterval:  30,
			ReconnectInterval: 15,
		}
		if err := e.CreateDevice(cfg); err != nil {
			t.Fatalf("CreateDevice %d failed: %v", i, err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	for i := 0; i < 3; i++ {
		if err := e.StartDevice(ctx, fmt.Sprintf("stopall-%d", i)); err != nil {
			t.Fatalf("StartDevice %d failed: %v", i, err)
		}
		time.Sleep(500 * time.Millisecond)
	}

	time.Sleep(3 * time.Second)

	stats := e.GetStats()
	if stats.OnlineDevices != 3 {
		t.Errorf("expected 3 online, got %d", stats.OnlineDevices)
	}

	e.StopAll()

	time.Sleep(1 * time.Second)

	stats = e.GetStats()
	if stats.OnlineDevices != 0 {
		t.Errorf("expected 0 online after StopAll, got %d", stats.OnlineDevices)
	}
}

// TestDeleteRunningDevice 测试删除运行中的设备
func TestDeleteRunningDevice(t *testing.T) {
	fp := newFakePlatform(t)
	defer fp.close()

	e := New()
	e.SetStorage(nil)

	cfg := &base.DeviceConfig{
		ID:                "del-test",
		Protocol:          "jt808",
		Phone:             "013800007777",
		TargetAddr:        fp.addr(),
		AuthCode:          "test",
		HeartbeatInterval: 60,
		LocationInterval:  30,
	}

	e.CreateDevice(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	e.StartDevice(ctx, "del-test")
	time.Sleep(3 * time.Second)

	if err := e.DeleteDevice("del-test"); err != nil {
		t.Fatalf("DeleteDevice failed: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	_, err := e.GetDevice("del-test")
	if err == nil {
		t.Error("expected error after delete, got nil")
	}

	stats := e.GetStats()
	if stats.TotalDevices != 0 {
		t.Errorf("expected TotalDevices=0, got %d", stats.TotalDevices)
	}
	if stats.OnlineDevices != 0 {
		t.Errorf("expected OnlineDevices=0, got %d", stats.OnlineDevices)
	}
}
