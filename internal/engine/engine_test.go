package engine

import (
	"testing"

	"github.com/suoten/jt-simulate/internal/simulator/base"
)

func TestCreateAndDeleteDevice(t *testing.T) {
	e := New()
	e.SetStorage(nil) // 不持久化

	cfg := &base.DeviceConfig{
		ID:         "test-001",
		Protocol:   "jt808",
		Phone:      "013800001111",
		Plate:      "测试A",
		PlateColor: 1,
		TargetAddr: "127.0.0.1:7611",
		AuthCode:   "test",
	}

	// 创建
	if err := e.CreateDevice(cfg); err != nil {
		t.Fatalf("CreateDevice failed: %v", err)
	}

	// 重复创建
	if err := e.CreateDevice(cfg); err == nil {
		t.Fatal("expected error on duplicate create, got nil")
	}

	// 查找
	info, err := e.GetDevice("test-001")
	if err != nil {
		t.Fatalf("GetDevice failed: %v", err)
	}
	if info.Config.Phone != "013800001111" {
		t.Errorf("unexpected phone: %s", info.Config.Phone)
	}
	if info.State() != base.StateOffline {
		t.Errorf("expected offline, got %s", info.State())
	}

	// 列表
	list := e.ListDevices()
	if len(list) != 1 {
		t.Errorf("expected 1 device, got %d", len(list))
	}

	// 统计
	stats := e.GetStats()
	if stats.TotalDevices != 1 {
		t.Errorf("expected TotalDevices=1, got %d", stats.TotalDevices)
	}
	if stats.OnlineDevices != 0 {
		t.Errorf("expected OnlineDevices=0, got %d", stats.OnlineDevices)
	}

	// 删除
	if err := e.DeleteDevice("test-001"); err != nil {
		t.Fatalf("DeleteDevice failed: %v", err)
	}

	// 删除后查找
	_, err = e.GetDevice("test-001")
	if err == nil {
		t.Fatal("expected error after delete, got nil")
	}

	// 统计归零
	stats = e.GetStats()
	if stats.TotalDevices != 0 {
		t.Errorf("expected TotalDevices=0 after delete, got %d", stats.TotalDevices)
	}
}

func TestStopAllEmpty(t *testing.T) {
	e := New()
	e.SetStorage(nil)
	// 空引擎 StopAll 不应 panic
	e.StopAll()
}

func TestDeviceStateAtomic(t *testing.T) {
	e := New()
	e.SetStorage(nil)

	cfg := &base.DeviceConfig{
		ID:         "state-test",
		Protocol:   "jt808",
		Phone:      "013800002222",
		TargetAddr: "127.0.0.1:7611",
	}
	e.CreateDevice(cfg)

	info, _ := e.GetDevice("state-test")

	// 并发读写测试
	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < 1000; i++ {
			_ = info.State()
		}
	}()

	for i := 0; i < 1000; i++ {
		info.setState(base.StateOnline)
		info.setState(base.StateOffline)
	}

	<-done
	// 如果没有 data race，测试通过
}
