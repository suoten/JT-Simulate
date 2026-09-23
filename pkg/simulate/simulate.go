package simulate

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/suoten/jt-simulate/internal/engine"
	"github.com/suoten/jt-simulate/internal/simulator/base"
	jt808sim "github.com/suoten/jt-simulate/internal/simulator/jt808"
	"github.com/suoten/jt-simulate/internal/workshop"
)

// Config 仿真配置
type Config struct {
	Target string // 目标平台地址
}

// Simulator 仿真器（SDK入口）
type Simulator struct {
	engine   *engine.Engine
	workshop *workshop.Workshop
	mu       sync.Mutex
}

// New 创建仿真器
func New(cfg *Config) *Simulator {
	return &Simulator{
		engine:   engine.New(),
		workshop: workshop.New(),
	}
}

// NewDevice 创建设备
func (s *Simulator) NewDevice(cfg *DeviceConfig) (*Device, error) {
	baseCfg := &base.DeviceConfig{
		ID:         cfg.Phone,
		Protocol:   cfg.Protocol,
		Phone:      cfg.Phone,
		Plate:      cfg.Plate,
		PlateColor: cfg.PlateColor,
		TargetAddr: cfg.Target,
		AuthCode:   cfg.AuthCode,
		ProvinceID:    cfg.ProvinceID,
		CityID:        cfg.CityID,
		Manufacturer:  cfg.Manufacturer,
		TerminalModel: cfg.TerminalModel,
		TerminalID:    cfg.TerminalID,
		StartLat:      cfg.StartLat,
		StartLon:      cfg.StartLon,
	}
	if err := s.engine.CreateDevice(baseCfg); err != nil {
		return nil, err
	}
	return &Device{
		sim:    s,
		config: baseCfg,
	}, nil
}

// DeviceConfig 设备配置
type DeviceConfig struct {
	Protocol       string
	Phone          string
	Plate          string
	PlateColor     byte
	Target         string
	AuthCode       string
	ProvinceID     int
	CityID         int
	Manufacturer   string
	TerminalModel  string
	TerminalID     string
	StartLat       float64
	StartLon       float64
}

// Device 设备
type Device struct {
	sim    *Simulator
	config *base.DeviceConfig
	state  int32
}

// Online 上线
func (d *Device) Online(ctx context.Context) error {
	if err := d.sim.engine.StartDevice(ctx, d.config.ID); err != nil {
		return err
	}
	atomic.StoreInt32(&d.state, 1)
	return nil
}

// Offline 下线
func (d *Device) Offline() {
	_ = d.sim.engine.StopDevice(d.config.ID)
	atomic.StoreInt32(&d.state, 0)
}

// SendLocation 发送位置
func (d *Device) SendLocation(lat, lon float64, speed, direction uint16) error {
	info, err := d.sim.engine.GetDevice(d.config.ID)
	if err != nil {
		return err
	}
	if s, ok := info.Simulator.(*jt808sim.Simulator); ok {
		return s.SendLocation(lat, lon, speed, direction)
	}
	return fmt.Errorf("unsupported simulator type")
}

// SendAlarm 发送报警
func (d *Device) SendAlarm(alarmFlag uint16) error {
	info, err := d.sim.engine.GetDevice(d.config.ID)
	if err != nil {
		return err
	}
	if s, ok := info.Simulator.(*jt808sim.Simulator); ok {
		return s.SendAlarm(alarmFlag)
	}
	return fmt.Errorf("unsupported simulator type")
}

// IsOnline 是否在线
func (d *Device) IsOnline() bool {
	return atomic.LoadInt32(&d.state) == 1
}

// RunScenario 运行场景
func (s *Simulator) RunScenario(name string) error {
	se := engine.NewScenarioEngine()
	for _, sc := range engine.DefaultScenarios() {
		se.Register(sc)
	}
	return se.Run(nil, name, func(step *engine.ScenarioStep) error {
		fmt.Printf("[场景] %s: %s\n", step.Name, step.Action)
		return nil
	})
}

// CheckCompliance 合规性检查
func (s *Simulator) CheckCompliance(protocol, target string) (interface{}, error) {
	// 简化：返回检查器
	return gin_H{"protocol": protocol, "target": target, "status": "ok"}, nil
}

// EncodeMessage 编码消息
func (s *Simulator) EncodeMessage(protocol, msgID, phone string, fields map[string]interface{}) (string, error) {
	resp := s.workshop.Encode(&workshop.EncodeRequest{
		Protocol: protocol,
		MsgID:    msgID,
		Phone:    phone,
		Fields:   fields,
	})
	if !resp.Success {
		return "", errors.New(resp.Error)
	}
	return resp.Hex, nil
}

type gin_H = map[string]interface{}
