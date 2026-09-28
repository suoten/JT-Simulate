package gbt32960

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/suoten/jt-simulate/internal/simulator/base"
	gbt32960 "github.com/suoten/jt-simulate/pkg/codec/gbt32960"
)

// Simulator GB/T 32960 仿真器（新能源汽车）
type Simulator struct {
	*base.Simulator
}

// New 创建 GBT32960 仿真器
func New(cfg *base.DeviceConfig) *Simulator {
	s := &Simulator{
		Simulator: base.NewSimulator(cfg),
	}
	// GBT32960 不需要处理平台下发消息（协议是单向上报）
	return s
}

// Online 上线（车辆登入→心跳→实时数据上报），带自动重连
func (s *Simulator) Online(ctx context.Context) error {
	if err := s.Connect(ctx); err != nil {
		return fmt.Errorf("connect: %w", err)
	}

	if err := s.sendLogin(); err != nil {
		return fmt.Errorf("vehicle login: %w", err)
	}

	// 启动心跳
	s.StartHeartbeat(ctx, s.sendHeartbeat)

	// 启动定时实时数据上报
	s.StartLocationReport(ctx, s.sendRealtimeInfo)

	// 启动自动重连
	s.StartReconnectLoop(ctx, func(ctx context.Context) error {
		if err := s.sendLogin(); err != nil {
			return err
		}
		s.StartHeartbeat(ctx, s.sendHeartbeat)
		s.StartLocationReport(ctx, s.sendRealtimeInfo)
		return nil
	})

	return nil
}

// Offline 下线
func (s *Simulator) Offline() {
	// 发送登出
	s.sendLogout()
	s.Disconnect()
}

// sendLogin 发送车辆登入
func (s *Simulator) sendLogin() error {
	cfg := s.Config()
	vin := cfg.Phone
	if vin == "" {
		vin = "LSGAB52A7DF123456"
	}
	now := time.Now().Format("060102150405")
	data := gbt32960.EncodeVehicleLogin(vin, now, "12345678901234567890", 1)
	return s.Send(data)
}

// sendLogout 发送车辆登出
func (s *Simulator) sendLogout() {
	cfg := s.Config()
	vin := cfg.Phone
	if vin == "" {
		vin = "LSGAB52A7DF123456"
	}
	now := time.Now().Format("060102150405")
	data := gbt32960.EncodeVehicleLogout(vin, now, 1)
	s.Send(data)
}

// sendHeartbeat 发送心跳
func (s *Simulator) sendHeartbeat() error {
	cfg := s.Config()
	vin := cfg.Phone
	if vin == "" {
		vin = "LSGAB52A7DF123456"
	}
	data := gbt32960.EncodeHeartbeat(vin)
	return s.Send(data)
}

// sendRealtimeInfo 发送实时数据上报
func (s *Simulator) sendRealtimeInfo() error {
	cfg := s.Config()
	vin := cfg.Phone
	if vin == "" {
		vin = "LSGAB52A7DF123456"
	}
	// 沿路线移动
	s.MoveAlongRoute()

	now := time.Now().Format("060102150405")

	vehData := &gbt32960.VehicleData{
		Status: 0x01, ChargeStatus: 0x01, Mode: 0x01,
		Speed:         uint16(s.CurSpeed()),
		TotalMileage:  100000,
		Voltage:       35000,
		Current:       5000,
		SOC:           80,
		DCDCStatus:    0x01,
		Shift:         0x02,
		Resistance:    500,
	}
	posData := &gbt32960.PositionData{
		ChargeStatus: 0x01,
		Lat:          int32(s.CurLat() * 1000000),
		Lon:          int32(s.CurLon() * 1000000),
		Altitude:     500,
		Direction:    s.CurDir(),
		Speed:        uint16(s.CurSpeed()),
	}
	items := []gbt32960.RealtimeDataItem{
		{Type: gbt32960.DataItemVehicle, Data: gbt32960.EncodeVehicleData(vehData)},
		{Type: gbt32960.DataItemPosition, Data: gbt32960.EncodePositionData(posData)},
	}
	data := gbt32960.EncodeRealtimeInfo(vin, now, 1, items)
	return s.Send(data)
}

// SendAlarm 发送故障报警
func (s *Simulator) SendAlarm(alarmFlag uint16) error {
	log.Printf("[%s] GBT32960 报警: 0x%04X", s.Config().Phone, alarmFlag)
	return nil
}
