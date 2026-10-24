package jt1045

import (
	"bytes"
	"testing"
)

// TestMarshalNoSideEffects Marshal 不得修改结构体的时间字段（前导零补齐只作用于局部副本）
func TestMarshalNoSideEffects(t *testing.T) {
	// ADASDataMessage.AlarmTime
	adas := &ADASDataMessage{
		VehicleColor: 1,
		VehiclePlate: "粤B12345",
		AlarmTime:    "2309231234", // 10 字节，需补零到 12
		AlarmType:    0x0101,
		AlarmLevel:   1,
		Lat:          22.5,
		Lon:          113.25,
		Speed:        600,
		Direction:    90,
		VehicleNo:    "粤B88888",
		VehicleSpeed: 500,
		Distance:     100,
	}
	b1, err := adas.Marshal()
	if err != nil {
		t.Fatalf("marshal adas data failed: %v", err)
	}
	if adas.AlarmTime != "2309231234" {
		t.Errorf("Marshal modified AlarmTime: got %q, want %q", adas.AlarmTime, "2309231234")
	}
	b2, _ := adas.Marshal()
	if !bytes.Equal(b1, b2) {
		t.Error("repeated Marshal produced different output for ADASDataMessage")
	}

	// TireAlarmMessage.AlarmTime
	tire := &TireAlarmMessage{
		VehicleColor: 1,
		VehiclePlate: "粤B12345",
		AlarmTime:    "230923123", // 9 字节，需补零到 12
		TirePosition: 1,
		Pressure:     2500,
		Temp:         35,
		Battery:      80,
		AlarmFlag:    0x01,
	}
	t1, err := tire.Marshal()
	if err != nil {
		t.Fatalf("marshal tire alarm failed: %v", err)
	}
	if len(t1) != 35 {
		t.Errorf("tire alarm marshal length = %d, want 35", len(t1))
	}
	if tire.AlarmTime != "230923123" {
		t.Errorf("Marshal modified AlarmTime: got %q, want %q", tire.AlarmTime, "230923123")
	}
	t2, _ := tire.Marshal()
	if !bytes.Equal(t1, t2) {
		t.Error("repeated Marshal produced different output for TireAlarmMessage")
	}
}
