package jt1253

import (
	"bytes"
	"testing"
)

// TestMarshalNoSideEffects Marshal 不得修改结构体的时间字段（前导零补齐只作用于局部副本）
func TestMarshalNoSideEffects(t *testing.T) {
	// HazardStatusMessage.StatusTime
	hs := &HazardStatusMessage{
		VehicleColor: 1,
		VehiclePlate: "粤B12345",
		StatusTime:   "2309231234", // 10 字节，需补零到 12
		HazardClass:  3,
		HazardState:  1,
		Lat:          22.5,
		Lon:          113.25,
		Speed:        600,
		Direction:    90,
		Temp:         25,
		Humidity:     60,
		Pressure:     1000,
		Level:        50,
	}
	b1, err := hs.Marshal()
	if err != nil {
		t.Fatalf("marshal hazard status failed: %v", err)
	}
	if hs.StatusTime != "2309231234" {
		t.Errorf("Marshal modified StatusTime: got %q, want %q", hs.StatusTime, "2309231234")
	}
	b2, _ := hs.Marshal()
	if !bytes.Equal(b1, b2) {
		t.Error("repeated Marshal produced different output for HazardStatusMessage")
	}

	// UnloadReportMessage.UnloadTime
	ur := &UnloadReportMessage{
		VehicleColor: 1,
		VehiclePlate: "粤B12345",
		UnloadTime:   "2309231234", // 10 字节，需补零到 12
		HazardClass:  3,
		UnloadAmount: 100,
		UnloadPlace:  "深圳",
		Lat:          22.5,
		Lon:          113.25,
	}
	u1, err := ur.Marshal()
	if err != nil {
		t.Fatalf("marshal unload report failed: %v", err)
	}
	if ur.UnloadTime != "2309231234" {
		t.Errorf("Marshal modified UnloadTime: got %q, want %q", ur.UnloadTime, "2309231234")
	}
	u2, _ := ur.Marshal()
	if !bytes.Equal(u1, u2) {
		t.Error("repeated Marshal produced different output for UnloadReportMessage")
	}
}
