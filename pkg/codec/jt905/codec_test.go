package jt905

import (
	"bytes"
	"testing"
)

// TestTaxiOperateUnmarshalBoundary 边界测试：
// 1. 恰好最小长度（51 字节）的数据能正确解析；
// 2. 少一字节（50 字节）返回错误而不是 panic；
// 3. 完整 59 字节（含可选的里程/金额字段）解析正确。
func TestTaxiOperateUnmarshalBoundary(t *testing.T) {
	orig := &TaxiOperateMessage{
		VehicleColor: 1,
		VehiclePlate: "粤B12345",
		OperateType:  1,
		GetOnTime:    "230923120000",
		GetOffTime:   "230923123000",
		GetOnLat:     22.5,
		GetOnLon:     113.25,
		GetOffLat:    22.75,
		GetOffLon:    113.5,
		Distance:     5000,
		Amount:       2500,
	}
	full, err := orig.Marshal()
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	if len(full) != 59 {
		t.Fatalf("full marshal length = %d, want 59", len(full))
	}

	// 恰好 51 字节：必选字段完整，可选的 distance/amount 缺失
	m := &TaxiOperateMessage{}
	if err := m.Unmarshal(full[:51]); err != nil {
		t.Fatalf("unmarshal 51-byte minimal data failed: %v", err)
	}
	if m.VehicleColor != orig.VehicleColor {
		t.Errorf("VehicleColor = %d, want %d", m.VehicleColor, orig.VehicleColor)
	}
	if m.VehiclePlate != orig.VehiclePlate {
		t.Errorf("VehiclePlate = %q, want %q", m.VehiclePlate, orig.VehiclePlate)
	}
	if m.OperateType != orig.OperateType {
		t.Errorf("OperateType = %d, want %d", m.OperateType, orig.OperateType)
	}
	if m.GetOnTime != orig.GetOnTime {
		t.Errorf("GetOnTime = %q, want %q", m.GetOnTime, orig.GetOnTime)
	}
	if m.GetOffTime != orig.GetOffTime {
		t.Errorf("GetOffTime = %q, want %q", m.GetOffTime, orig.GetOffTime)
	}
	if m.GetOnLat != 22.5 || m.GetOnLon != 113.25 {
		t.Errorf("GetOn coords = %v/%v, want 22.5/113.25", m.GetOnLat, m.GetOnLon)
	}
	if m.GetOffLat != 22.75 || m.GetOffLon != 113.5 {
		t.Errorf("GetOff coords = %v/%v, want 22.75/113.5", m.GetOffLat, m.GetOffLon)
	}
	if m.Distance != 0 || m.Amount != 0 {
		t.Errorf("optional fields should stay zero on minimal data, got %d/%d", m.Distance, m.Amount)
	}

	// 少一字节（50 字节）：必须返回错误而不是 panic
	m2 := &TaxiOperateMessage{}
	if err := m2.Unmarshal(full[:50]); err == nil {
		t.Error("unmarshal 50-byte data should return error, got nil")
	}

	// 完整 59 字节
	m3 := &TaxiOperateMessage{}
	if err := m3.Unmarshal(full); err != nil {
		t.Fatalf("unmarshal full data failed: %v", err)
	}
	if m3.Distance != orig.Distance {
		t.Errorf("Distance = %d, want %d", m3.Distance, orig.Distance)
	}
	if m3.Amount != orig.Amount {
		t.Errorf("Amount = %d, want %d", m3.Amount, orig.Amount)
	}
	if m3.GetOffLon != orig.GetOffLon {
		t.Errorf("GetOffLon = %v, want %v", m3.GetOffLon, orig.GetOffLon)
	}
}

// TestMarshalNoSideEffects Marshal 不得修改结构体的时间字段（前导零补齐只作用于局部副本）
func TestMarshalNoSideEffects(t *testing.T) {
	// TaxiStatusMessage.ChangeTime
	st := &TaxiStatusMessage{
		VehicleColor: 1,
		VehiclePlate: "粤B12345",
		Status:       1,
		ChangeTime:   "2309231234", // 10 字节，需补零到 12
		Lat:          22.5,
		Lon:          113.25,
	}
	b1, err := st.Marshal()
	if err != nil {
		t.Fatalf("marshal status failed: %v", err)
	}
	if st.ChangeTime != "2309231234" {
		t.Errorf("Marshal modified ChangeTime: got %q, want %q", st.ChangeTime, "2309231234")
	}
	b2, _ := st.Marshal()
	if !bytes.Equal(b1, b2) {
		t.Error("repeated Marshal produced different output for TaxiStatusMessage")
	}

	// TaxiPriceMessage.StartTime
	pr := &TaxiPriceMessage{
		StartTime:      "2309231234",
		UnitPrice:      1000,
		UnitDistance:   2,
		PerKmPrice:     250,
		NightSurcharge: 50,
	}
	p1, err := pr.Marshal()
	if err != nil {
		t.Fatalf("marshal price failed: %v", err)
	}
	if pr.StartTime != "2309231234" {
		t.Errorf("Marshal modified StartTime: got %q, want %q", pr.StartTime, "2309231234")
	}
	p2, _ := pr.Marshal()
	if !bytes.Equal(p1, p2) {
		t.Error("repeated Marshal produced different output for TaxiPriceMessage")
	}

	// TaxiOperateMessage 的两个时间字段
	op := &TaxiOperateMessage{
		VehicleColor: 1,
		VehiclePlate: "粤B12345",
		OperateType:  1,
		GetOnTime:    "23092312",
		GetOffTime:   "23092313",
		GetOnLat:     22.5,
		GetOnLon:     113.25,
		GetOffLat:    22.75,
		GetOffLon:    113.5,
		Distance:     100,
		Amount:       200,
	}
	o1, err := op.Marshal()
	if err != nil {
		t.Fatalf("marshal operate failed: %v", err)
	}
	if op.GetOnTime != "23092312" || op.GetOffTime != "23092313" {
		t.Errorf("Marshal modified operate times: got %q/%q", op.GetOnTime, op.GetOffTime)
	}
	o2, _ := op.Marshal()
	if !bytes.Equal(o1, o2) {
		t.Error("repeated Marshal produced different output for TaxiOperateMessage")
	}
}
