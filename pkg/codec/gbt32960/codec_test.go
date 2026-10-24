package gbt32960

import (
	"bytes"
	"encoding/binary"
	"testing"
)

// TestEncodeFrameLayoutGB 验证报文帧结构符合 GB/T 32960.3-2016：
// 起始符 2 字节 0x23 0x23，头总长 24 字节，各字段落在国标偏移上
func TestEncodeFrameLayoutGB(t *testing.T) {
	vin := "LSGAB52A7DF123456"
	body := []byte{0x26, 0x10, 24, 26, 10, 14, 0x00, 0x01}
	h := &Header{
		CmdFlag:     MsgTypeVehicleLogin,
		RespFlag:    0xFE,
		VIN:         vin,
		EncryptType: 0x01,
	}
	data := Encode(h, body)

	if len(data) != HeaderLen+len(body)+1 {
		t.Fatalf("frame length = %d, want %d (24 header + %d body + 1 BCC)", len(data), HeaderLen+len(body)+1, len(body))
	}
	if data[0] != 0x23 || data[1] != 0x23 {
		t.Errorf("start chars = % X, want 23 23", data[0:2])
	}
	if data[2] != MsgTypeVehicleLogin {
		t.Errorf("cmd flag at offset 2 = 0x%02X", data[2])
	}
	if data[3] != 0xFE {
		t.Errorf("resp flag at offset 3 = 0x%02X", data[3])
	}
	if gotVIN := string(data[4:21]); gotVIN != vin {
		t.Errorf("VIN at offsets 4..21 = %q, want %q", gotVIN, vin)
	}
	if data[21] != 0x01 {
		t.Errorf("encrypt type at offset 21 = 0x%02X", data[21])
	}
	if gotLen := binary.BigEndian.Uint16(data[22:24]); gotLen != uint16(len(body)) {
		t.Errorf("body length at offsets 22..24 = %d, want %d", gotLen, len(body))
	}
	if !bytes.Equal(data[24:24+len(body)], body) {
		t.Errorf("body at offset 24 mismatch")
	}

	// BCC = 命令标识(偏移2)到数据单元末尾的异或，不含起始符与 BCC
	var want byte
	for _, b := range data[2 : 24+len(body)] {
		want ^= b
	}
	if data[len(data)-1] != want {
		t.Errorf("BCC = 0x%02X, want 0x%02X", data[len(data)-1], want)
	}
}

// TestDecodeRejectsSingleStartChar 单起始符（旧版错位帧）应被拒绝
func TestDecodeRejectsSingleStartChar(t *testing.T) {
	h := &Header{CmdFlag: MsgTypeHeartbeat, RespFlag: 0xFE, VIN: "LSGAB52A7DF123456", EncryptType: 0x01}
	frame := Encode(h, nil)

	// 构造单起始符帧：去掉第二个 0x23
	bad := append([]byte{0x23}, frame[2:]...)
	if _, err := Decode(bad); err == nil {
		t.Error("Decode accepted a frame with a single start char (off-by-one header)")
	}
}

// TestDecodeBCCError 篡改任一字节应导致 BCC 校验失败
func TestDecodeBCCError(t *testing.T) {
	h := &Header{CmdFlag: MsgTypeHeartbeat, RespFlag: 0xFE, VIN: "LSGAB52A7DF123456", EncryptType: 0x01}
	frame := Encode(h, nil)
	frame[5] ^= 0xFF
	if _, err := Decode(frame); err == nil {
		t.Error("Decode accepted a corrupted frame (BCC should mismatch)")
	}
}

// TestRoundTripVehicleLogin 车辆登入 round-trip
func TestRoundTripVehicleLogin(t *testing.T) {
	vin := "LSGAB52A7DF123456"
	data := EncodeVehicleLogin(vin, "260923112233", "12345678901234567890", 7)

	pkt, err := Decode(data)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	if pkt.Header.VIN != vin {
		t.Errorf("VIN = %q, want %q", pkt.Header.VIN, vin)
	}
	if pkt.Header.CmdFlag != MsgTypeVehicleLogin {
		t.Errorf("cmd flag = 0x%02X", pkt.Header.CmdFlag)
	}
	if pkt.Header.BodyLen != uint16(len(pkt.Body)) {
		t.Errorf("body len = %d, actual %d", pkt.Header.BodyLen, len(pkt.Body))
	}
	// 登入体：时间6BCD + 序号2 + ICCID 20 + ...
	if got := binary.BigEndian.Uint16(pkt.Body[6:8]); got != 7 {
		t.Errorf("login seq = %d, want 7", got)
	}
	if string(pkt.Body[8:28]) != "12345678901234567890" {
		t.Errorf("ICCID = %q", string(pkt.Body[8:28]))
	}
}

// TestPositionDataLayoutGB 位置数据：经度在前、纬度在后，共 15 字节
func TestPositionDataLayoutGB(t *testing.T) {
	d := &PositionData{
		LocStatus: 0x00,
		Lon:       116397440, // 116.397440°E
		Lat:       39909300,  // 39.909300°N
		Altitude:  5000,
		Direction: 180,
		Speed:     600,
	}
	b := EncodePositionData(d)
	if len(b) != 15 {
		t.Fatalf("position data length = %d, want 15", len(b))
	}
	if b[0] != 0x00 {
		t.Errorf("loc status = 0x%02X, want 0x00", b[0])
	}
	if got := int32(binary.BigEndian.Uint32(b[1:5])); got != d.Lon {
		t.Errorf("longitude at offsets 1..5 = %d, want %d (经度在前)", got, d.Lon)
	}
	if got := int32(binary.BigEndian.Uint32(b[5:9])); got != d.Lat {
		t.Errorf("latitude at offsets 5..9 = %d, want %d (纬度在后)", got, d.Lat)
	}
	if got := binary.BigEndian.Uint16(b[9:11]); got != 5000 {
		t.Errorf("altitude = %d", got)
	}
	if got := binary.BigEndian.Uint16(b[11:13]); got != 180 {
		t.Errorf("direction = %d", got)
	}
	if got := binary.BigEndian.Uint16(b[13:15]); got != 600 {
		t.Errorf("speed = %d", got)
	}
}

// TestRoundTripRealtimeInfo 实时信息上报 round-trip（含帧级 + 数据项）
func TestRoundTripRealtimeInfo(t *testing.T) {
	vin := "LSGAB52A7DF123456"
	veh := &VehicleData{
		Status: 0x01, ChargeStatus: 0x01, Mode: 0x01,
		Speed: 800, TotalMileage: 100000, Voltage: 35000, Current: 5000,
		SOC: 80, DCDCStatus: 0x01, Shift: 0x02, Resistance: 500,
	}
	pos := &PositionData{LocStatus: 0x00, Lon: 116397440, Lat: 39909300, Altitude: 5000, Direction: 90, Speed: 800}
	items := []RealtimeDataItem{
		{Type: DataItemVehicle, Data: EncodeVehicleData(veh)},
		{Type: DataItemPosition, Data: EncodePositionData(pos)},
	}
	data := EncodeRealtimeInfo(vin, "260923112233", 3, items)

	pkt, err := Decode(data)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	if pkt.Header.CmdFlag != MsgTypeRealtimeInfo {
		t.Errorf("cmd flag = 0x%02X", pkt.Header.CmdFlag)
	}
	// 实时信息体：时间6 + 序号2 + 项数2 + 数据项
	if got := binary.BigEndian.Uint16(pkt.Body[6:8]); got != 3 {
		t.Errorf("seq = %d, want 3", got)
	}
	if got := binary.BigEndian.Uint16(pkt.Body[8:10]); got != 2 {
		t.Errorf("item count = %d, want 2", got)
	}
	// 第一项：类型1 + 长度2 + 数据
	if pkt.Body[10] != DataItemVehicle {
		t.Errorf("item1 type = 0x%02X", pkt.Body[10])
	}
	vehLen := int(binary.BigEndian.Uint16(pkt.Body[11:13]))
	if vehLen != 18 {
		t.Errorf("vehicle data length = %d, want 18", vehLen)
	}
}

// TestHeartbeatAndLogout 心跳/登出帧 round-trip
func TestHeartbeatAndLogout(t *testing.T) {
	vin := "LSGAB52A7DF123456"
	hb := EncodeHeartbeat(vin)
	pkt, err := Decode(hb)
	if err != nil {
		t.Fatalf("heartbeat decode failed: %v", err)
	}
	if pkt.Header.CmdFlag != MsgTypeHeartbeat || pkt.Header.BodyLen != 0 {
		t.Errorf("heartbeat header mismatch: cmd=0x%02X bodyLen=%d", pkt.Header.CmdFlag, pkt.Header.BodyLen)
	}

	lo := EncodeVehicleLogout(vin, "260923112233", 9)
	pkt2, err := Decode(lo)
	if err != nil {
		t.Fatalf("logout decode failed: %v", err)
	}
	if got := binary.BigEndian.Uint16(pkt2.Body[6:8]); got != 9 {
		t.Errorf("logout seq = %d, want 9", got)
	}
}
