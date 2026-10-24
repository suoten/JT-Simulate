package jt809

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func TestEncodeVersionDifference(t *testing.T) {
	// 2011 版：VersionFlag=0，头长 22 字节
	h2011 := &Header{
		MsgSN:           1,
		MsgID:           MsgIDUpLinktestReq, // 0x1005，无消息体
		MsgGNSSCENTERID: 1001,
		VersionFlag:     0, // 2011
	}
	data2011, err := NewCodec().Encode(h2011, nil)
	if err != nil {
		t.Fatalf("encode 2011 failed: %v", err)
	}

	// 2019 版：VersionFlag=1，头长 30 字节
	// 固定 UTCTime（8 字节内不含 0x5A/0x5B/0x5D/0x5E），
	// 避免时间戳中的特殊字节被转义后帧长度随时间波动
	h2019 := &Header{
		MsgSN:           1,
		MsgID:           MsgIDUpLinktestReq,
		MsgGNSSCENTERID: 1001,
		VersionFlag:     1, // 2019
		UTCTime:         1700000001000,
	}
	data2019, err := NewCodec().Encode(h2019, nil)
	if err != nil {
		t.Fatalf("encode 2019 failed: %v", err)
	}

	t.Logf("2011 hex: %s", hex.EncodeToString(data2011))
	t.Logf("2019 hex: %s", hex.EncodeToString(data2019))

	// 2011 版应该比 2019 版短 8 字节（头长度差异：30-22=8）
	// 2011: 0x5B + header(22) + CRC16(2) + 0x5D = 26
	// 2019: 0x5B + header(30) + CRC16(2) + 0x5D = 34
	if len(data2011) != 26 {
		t.Errorf("2011 frame length = %d, want 26", len(data2011))
	}
	if len(data2019) != 34 {
		t.Errorf("2019 frame length = %d, want 34", len(data2019))
	}

	// 两个报文必须不同
	if hex.EncodeToString(data2011) == hex.EncodeToString(data2019) {
		t.Error("2011 and 2019 frames should be different")
	}

	// 验证首尾分隔符：帧首 0x5B，帧尾 0x5D
	if data2011[0] != BeginFlag || data2011[len(data2011)-1] != EndFlag {
		t.Errorf("2011 frame should start with 0x5B and end with 0x5D, got 0x%02X...0x%02X", data2011[0], data2011[len(data2011)-1])
	}
	if data2019[0] != BeginFlag || data2019[len(data2019)-1] != EndFlag {
		t.Errorf("2019 frame should start with 0x5B and end with 0x5D, got 0x%02X...0x%02X", data2019[0], data2019[len(data2019)-1])
	}
}

func TestDecodeVersionAutoDetect(t *testing.T) {
	// 编码 2011 版
	h2011 := &Header{
		MsgSN:           1,
		MsgID:           MsgIDUpLinktestReq,
		MsgGNSSCENTERID: 1001,
		VersionFlag:     0,
	}
	data2011, _ := NewCodec().Encode(h2011, nil)

	// 解码 2011 版
	h, body, err := NewCodec().Decode(data2011)
	if err != nil {
		t.Fatalf("decode 2011 failed: %v", err)
	}
	if h.VersionFlag != 0 {
		t.Errorf("decoded VersionFlag = %d, want 0 (2011)", h.VersionFlag)
	}
	if len(body) != 0 {
		t.Errorf("2011 body length = %d, want 0", len(body))
	}

	// 编码 2019 版
	h2019 := &Header{
		MsgSN:           1,
		MsgID:           MsgIDUpLinktestReq,
		MsgGNSSCENTERID: 1001,
		VersionFlag:     1,
	}
	data2019, _ := NewCodec().Encode(h2019, nil)

	// 解码 2019 版
	h, body, err = NewCodec().Decode(data2019)
	if err != nil {
		t.Fatalf("decode 2019 failed: %v", err)
	}
	if h.VersionFlag != 1 {
		t.Errorf("decoded VersionFlag = %d, want 1 (2019)", h.VersionFlag)
	}
	if len(body) != 0 {
		t.Errorf("2019 body length = %d, want 0", len(body))
	}
}

func TestEncodeWithBody(t *testing.T) {
	// 带消息体的编码测试（位置上报）
	loc := &LocationMessage{
		VehicleColor: 1,
		VehiclePlate: "京A12345",
		Lat:          39.9093,
		Lon:          116.3974,
		Speed:        600,
		Direction:    180,
		Time:         "230101120000",
	}
	bodyBytes, err := loc.Marshal()
	if err != nil {
		t.Fatalf("marshal location failed: %v", err)
	}

	// 2011 版
	h2011 := &Header{
		MsgSN:           1,
		MsgID:           MsgIDLocationMsg,
		MsgGNSSCENTERID: 1001,
		VersionFlag:     0,
	}
	data2011, _ := NewCodec().Encode(h2011, bodyBytes)

	// 2019 版（固定 UTCTime，其 8 字节不含 0x5A/0x5B/0x5D/0x5E，
	// 确保转义不改变帧长度，长度断言稳定）
	h2019 := &Header{
		MsgSN:           1,
		MsgID:           MsgIDLocationMsg,
		MsgGNSSCENTERID: 1001,
		VersionFlag:     1,
		UTCTime:         1700000001000,
	}
	data2019, _ := NewCodec().Encode(h2019, bodyBytes)

	t.Logf("2011 location hex: %s", hex.EncodeToString(data2011))
	t.Logf("2019 location hex: %s", hex.EncodeToString(data2019))

	// 长度差 = 8（头长度差异：30-22=8）
	if len(data2019)-len(data2011) != 8 {
		t.Errorf("length diff = %d, want 8", len(data2019)-len(data2011))
	}

	// 解码验证
	h, body, err := NewCodec().Decode(data2011)
	if err != nil {
		t.Fatalf("decode 2011 location failed: %v", err)
	}
	if h.VersionFlag != 0 {
		t.Errorf("2011 VersionFlag = %d, want 0", h.VersionFlag)
	}
	loc2 := &LocationMessage{}
	if err := loc2.Unmarshal(body); err != nil {
		t.Errorf("unmarshal 2011 location failed: %v", err)
	}

	h, body, err = NewCodec().Decode(data2019)
	if err != nil {
		t.Fatalf("decode 2019 location failed: %v", err)
	}
	if h.VersionFlag != 1 {
		t.Errorf("2019 VersionFlag = %d, want 1", h.VersionFlag)
	}
	loc3 := &LocationMessage{}
	if err := loc3.Unmarshal(body); err != nil {
		t.Errorf("unmarshal 2019 location failed: %v", err)
	}
}

func TestFrameDelimiters(t *testing.T) {
	h := &Header{MsgSN: 1, MsgID: MsgIDUpLinktestReq, MsgGNSSCENTERID: 1001, VersionFlag: 1}
	data, _ := NewCodec().Encode(h, nil)

	// JT/T 809 帧首 = 0x5B, 帧尾 = 0x5D
	if data[0] != BeginFlag {
		t.Errorf("frame start = 0x%02X, want 0x5B", data[0])
	}
	if data[len(data)-1] != EndFlag {
		t.Errorf("frame end = 0x%02X, want 0x5D", data[len(data)-1])
	}
	t.Logf("frame: %s", fmt.Sprintf("% X", data))
}

func TestEscapeAndUnescape(t *testing.T) {
	// 测试转义：0x5B 和 0x5D 在数据中应被正确转义和反转义
	// 使用包含 0x5B 和 0x5D 的消息体
	loc := &LocationMessage{
		VehicleColor: 1,
		VehiclePlate: "京A12345",
		Lat:          39.9093,
		Lon:          116.3974,
		Speed:        600,
		Direction:    180,
		Time:         "230101120000",
	}
	bodyBytes, _ := loc.Marshal()

	h := &Header{
		MsgSN:           1,
		MsgID:           MsgIDLocationMsg,
		MsgGNSSCENTERID: 1001,
		VersionFlag:     1,
	}
	data, _ := NewCodec().Encode(h, bodyBytes)

	// 确保帧尾是 0x5D
	if data[len(data)-1] != EndFlag {
		t.Errorf("frame end = 0x%02X, want 0x5D", data[len(data)-1])
	}

	// 解码应该成功
	h2, body2, err := NewCodec().Decode(data)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if h2.MsgID != MsgIDLocationMsg {
		t.Errorf("MsgID = 0x%04X, want 0x%04X", h2.MsgID, MsgIDLocationMsg)
	}

	loc2 := &LocationMessage{}
	if err := loc2.Unmarshal(body2); err != nil {
		t.Fatalf("unmarshal location failed: %v", err)
	}
	if loc2.VehiclePlate != loc.VehiclePlate {
		t.Errorf("plate = %q, want %q", loc2.VehiclePlate, loc.VehiclePlate)
	}
}
