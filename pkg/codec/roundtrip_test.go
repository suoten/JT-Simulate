package codec_test

import (
	"encoding/hex"
	"testing"

	"github.com/suoten/jt-simulate/pkg/codec/gbt32960"
	"github.com/suoten/jt-simulate/pkg/codec/jt1045"
	"github.com/suoten/jt-simulate/pkg/codec/jt1078"
	"github.com/suoten/jt-simulate/pkg/codec/jt1253"
	"github.com/suoten/jt-simulate/pkg/codec/jt808"
	"github.com/suoten/jt-simulate/pkg/codec/jt809"
	"github.com/suoten/jt-simulate/pkg/codec/jt905"
	"github.com/suoten/jt-simulate/pkg/types"
)

// ============================================================
// JT/T 808 Round-trip Tests
// ============================================================

func TestJT808_RoundTrip_Location(t *testing.T) {
	codec := jt808.NewCodec()
	orig := &jt808.LocationMessage{
		AlarmFlag:  0x00000001,
		StatusFlag: 0x00000002,
		Latitude:   22.543210,
		Longitude:  113.945678,
		Altitude:   100,
		Speed:      600,
		Direction:  180,
		Time:       "230923123456",
	}
	header := &types.MessageHeader{
		MsgID: jt808.MsgIDLocation, Phone: "13800138000", SeqNum: 1,
		Version2019: true, ProtocolVer: 1,
	}
	data, err := codec.Encode(header, orig)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}
	t.Logf("encoded: %s", hex.EncodeToString(data))

	msg, err := codec.Decode(data)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	loc, ok := msg.Body.(*jt808.LocationMessage)
	if !ok {
		t.Fatalf("expected *LocationMessage, got %T", msg.Body)
	}
	if loc.Latitude != orig.Latitude {
		t.Errorf("latitude mismatch: got %f, want %f", loc.Latitude, orig.Latitude)
	}
	if loc.Longitude != orig.Longitude {
		t.Errorf("longitude mismatch: got %f, want %f", loc.Longitude, orig.Longitude)
	}
	if loc.AlarmFlag != orig.AlarmFlag {
		t.Errorf("alarm_flag mismatch: got %d, want %d", loc.AlarmFlag, orig.AlarmFlag)
	}
	if loc.Speed != orig.Speed {
		t.Errorf("speed mismatch: got %d, want %d", loc.Speed, orig.Speed)
	}
	if loc.Time != orig.Time {
		t.Errorf("time mismatch: got %s, want %s", loc.Time, orig.Time)
	}
}

func TestJT808_RoundTrip_Register(t *testing.T) {
	codec := jt808.NewCodec()
	orig := &jt808.RegisterMessage{
		ProvinceID:    4400,
		CityID:        4403,
		Manufacturer:  "SUOTE",
		TerminalModel: "JT-100",
		TerminalID:    "ABC123",
		PlateColor:    1,
		PlateNumber:   "粤B12345",
	}
	header := &types.MessageHeader{
		MsgID: jt808.MsgIDRegister, Phone: "13800138000", SeqNum: 1,
		Version2019: true, ProtocolVer: 1,
	}
	data, err := codec.Encode(header, orig)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}
	t.Logf("encoded: %s", hex.EncodeToString(data))

	msg, err := codec.Decode(data)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	reg, ok := msg.Body.(*jt808.RegisterMessage)
	if !ok {
		t.Fatalf("expected *RegisterMessage, got %T", msg.Body)
	}
	if reg.Manufacturer != orig.Manufacturer {
		t.Errorf("manufacturer mismatch: got %q, want %q", reg.Manufacturer, orig.Manufacturer)
	}
	if reg.TerminalID != orig.TerminalID {
		t.Errorf("terminal_id mismatch: got %q, want %q", reg.TerminalID, orig.TerminalID)
	}
	if reg.ProvinceID != orig.ProvinceID {
		t.Errorf("province_id mismatch: got %d, want %d", reg.ProvinceID, orig.ProvinceID)
	}
}

func TestJT808_RoundTrip_Heartbeat(t *testing.T) {
	codec := jt808.NewCodec()
	orig := &jt808.HeartbeatMessage{}
	header := &types.MessageHeader{
		MsgID: jt808.MsgIDHeartbeat, Phone: "13800138000", SeqNum: 1,
		Version2019: true, ProtocolVer: 1,
	}
	data, err := codec.Encode(header, orig)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}
	msg, err := codec.Decode(data)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if _, ok := msg.Body.(*jt808.HeartbeatMessage); !ok {
		t.Fatalf("expected *HeartbeatMessage, got %T", msg.Body)
	}
}

func TestJT808_RoundTrip_Auth(t *testing.T) {
	codec := jt808.NewCodec()
	orig := &jt808.AuthMessage{
		AuthCode: "AUTH123456",
		IMEI:     "123456789012345",
	}
	header := &types.MessageHeader{
		MsgID: jt808.MsgIDAuth, Phone: "13800138000", SeqNum: 1,
		Version2019: true, ProtocolVer: 1,
	}
	data, err := codec.Encode(header, orig)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}
	msg, err := codec.Decode(data)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	auth, ok := msg.Body.(*jt808.AuthMessage)
	if !ok {
		t.Fatalf("expected *AuthMessage, got %T", msg.Body)
	}
	if auth.AuthCode != orig.AuthCode {
		t.Errorf("auth_code mismatch: got %q, want %q", auth.AuthCode, orig.AuthCode)
	}
	if auth.IMEI != orig.IMEI {
		t.Errorf("imei mismatch: got %q, want %q", auth.IMEI, orig.IMEI)
	}
}

func TestJT808_RoundTrip_TextSend(t *testing.T) {
	codec := jt808.NewCodec()
	orig := &jt808.TextSendMessage{Flag: 1, Content: "Hello World"}
	header := &types.MessageHeader{
		MsgID: jt808.MsgIDTextSend, Phone: "13800138000", SeqNum: 1,
		Version2019: true, ProtocolVer: 1,
	}
	data, err := codec.Encode(header, orig)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}
	msg, err := codec.Decode(data)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	txt, ok := msg.Body.(*jt808.TextSendMessage)
	if !ok {
		t.Fatalf("expected *TextSendMessage, got %T", msg.Body)
	}
	if txt.Content != orig.Content {
		t.Errorf("content mismatch: got %q, want %q", txt.Content, orig.Content)
	}
}

func TestJT808_RoundTrip_Location_2011(t *testing.T) {
	codec := jt808.NewCodec()
	orig := &jt808.LocationMessage{
		AlarmFlag:  0x00000001,
		StatusFlag: 0x00000002,
		Latitude:   22.543210,
		Longitude:  113.945678,
		Altitude:   100,
		Speed:      600,
		Direction:  180,
		Time:       "230923123456",
	}
	header := &types.MessageHeader{
		MsgID: jt808.MsgIDLocation, Phone: "13800138000", SeqNum: 1,
		Version2019: false,
	}
	data, err := codec.Encode(header, orig)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}
	msg, err := codec.Decode(data)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	loc, ok := msg.Body.(*jt808.LocationMessage)
	if !ok {
		t.Fatalf("expected *LocationMessage, got %T", msg.Body)
	}
	if loc.Latitude != orig.Latitude {
		t.Errorf("latitude mismatch: got %f, want %f", loc.Latitude, orig.Latitude)
	}
}

// ============================================================
// JT/T 809 Round-trip Tests
// ============================================================

func TestJT809_RoundTrip_ConnectReq(t *testing.T) {
	codec := jt809.NewCodec()
	orig := &jt809.ConnectReqMessage{
		UserName:     1001,
		Password:     "password",
		DownLinkIP:   "127.0.0.1",
		DownLinkPort: 7611,
	}
	bodyBytes, err := orig.Marshal()
	if err != nil {
		t.Fatalf("marshal body failed: %v", err)
	}
	h := &jt809.Header{MsgSN: 1, MsgID: jt809.MsgIDUpConnectReq, GNSSCenterID: 1001, VersionFlag: 1}
	data, err := codec.Encode(h, bodyBytes)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}
	t.Logf("encoded: %s", hex.EncodeToString(data))

	decH, decBody, err := codec.Decode(data)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if decH.MsgID != jt809.MsgIDUpConnectReq {
		t.Errorf("msg_id mismatch: got 0x%04X, want 0x%04X", decH.MsgID, jt809.MsgIDUpConnectReq)
	}
	dec := &jt809.ConnectReqMessage{}
	if err := dec.Unmarshal(decBody); err != nil {
		t.Fatalf("unmarshal body failed: %v", err)
	}
	if dec.UserName != orig.UserName {
		t.Errorf("username mismatch: got %d, want %d", dec.UserName, orig.UserName)
	}
	if dec.Password != orig.Password {
		t.Errorf("password mismatch: got %q, want %q", dec.Password, orig.Password)
	}
	if dec.DownLinkIP != orig.DownLinkIP {
		t.Errorf("downlink_ip mismatch: got %q, want %q", dec.DownLinkIP, orig.DownLinkIP)
	}
	if dec.DownLinkPort != orig.DownLinkPort {
		t.Errorf("downlink_port mismatch: got %d, want %d", dec.DownLinkPort, orig.DownLinkPort)
	}
}

func TestJT809_RoundTrip_Location(t *testing.T) {
	codec := jt809.NewCodec()
	orig := &jt809.LocationMessage{
		VehicleColor: 1,
		VehiclePlate: "粤B12345",
		AlarmFlag:    0,
		Lat:          22.543210,
		Lon:          113.945678,
		Speed:        600,
		Direction:    180,
		Time:         "230923123456",
	}
	bodyBytes, err := orig.Marshal()
	if err != nil {
		t.Fatalf("marshal body failed: %v", err)
	}
	h := &jt809.Header{MsgSN: 1, MsgID: jt809.MsgIDLocationMsg, GNSSCenterID: 1001, VersionFlag: 1}
	data, err := codec.Encode(h, bodyBytes)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	decH, decBody, err := codec.Decode(data)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if decH.MsgID != jt809.MsgIDLocationMsg {
		t.Errorf("msg_id mismatch: got 0x%04X", decH.MsgID)
	}
	dec := &jt809.LocationMessage{}
	if err := dec.Unmarshal(decBody); err != nil {
		t.Fatalf("unmarshal body failed: %v", err)
	}
	if dec.VehiclePlate != orig.VehiclePlate {
		t.Errorf("plate mismatch: got %q, want %q", dec.VehiclePlate, orig.VehiclePlate)
	}
	if dec.Lat != orig.Lat {
		t.Errorf("lat mismatch: got %f, want %f", dec.Lat, orig.Lat)
	}
	if dec.Lon != orig.Lon {
		t.Errorf("lon mismatch: got %f, want %f", dec.Lon, orig.Lon)
	}
}

func TestJT809_RoundTrip_Alarm(t *testing.T) {
	codec := jt809.NewCodec()
	orig := &jt809.AlarmMessage{
		VehicleColor: 1,
		VehiclePlate: "粤B12345",
		AlarmSource:  1,
		AlarmType:    0x0101,
		AlarmTime:    "230923123456",
		Lat:          22.543210,
		Lon:          113.945678,
		Speed:        600,
		Direction:    180,
	}
	bodyBytes, _ := orig.Marshal()
	h := &jt809.Header{MsgSN: 1, MsgID: jt809.MsgIDAlarmMsg, GNSSCenterID: 1001, VersionFlag: 1}
	data, err := codec.Encode(h, bodyBytes)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	_, decBody, err := codec.Decode(data)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	dec := &jt809.AlarmMessage{}
	if err := dec.Unmarshal(decBody); err != nil {
		t.Fatalf("unmarshal body failed: %v", err)
	}
	if dec.VehiclePlate != orig.VehiclePlate {
		t.Errorf("plate mismatch: got %q, want %q", dec.VehiclePlate, orig.VehiclePlate)
	}
	if dec.AlarmType != orig.AlarmType {
		t.Errorf("alarm_type mismatch: got 0x%04X, want 0x%04X", dec.AlarmType, orig.AlarmType)
	}
	if dec.Lat != orig.Lat {
		t.Errorf("lat mismatch: got %f, want %f", dec.Lat, orig.Lat)
	}
}

func TestJT809_RoundTrip_LinktestReq(t *testing.T) {
	codec := jt809.NewCodec()
	h := &jt809.Header{MsgSN: 1, MsgID: jt809.MsgIDUpLinktestReq, GNSSCenterID: 1001, VersionFlag: 1}
	data, err := codec.Encode(h, nil)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}
	decH, _, err := codec.Decode(data)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if decH.MsgID != jt809.MsgIDUpLinktestReq {
		t.Errorf("msg_id mismatch: got 0x%04X", decH.MsgID)
	}
}

func TestJT809_RoundTrip_Statistics(t *testing.T) {
	codec := jt809.NewCodec()
	orig := &jt809.StatisticsMessage{
		VehicleColor:   1,
		VehiclePlate:   "粤B12345",
		TotalMsgCount:  100,
		LocationCount:  80,
		AlarmCount:     5,
		OnlineDuration: 3600,
	}
	bodyBytes, _ := orig.Marshal()
	h := &jt809.Header{MsgSN: 1, MsgID: jt809.MsgIDStatisticsMsg, GNSSCenterID: 1001, VersionFlag: 1}
	data, err := codec.Encode(h, bodyBytes)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	_, decBody, err := codec.Decode(data)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	dec := &jt809.StatisticsMessage{}
	if err := dec.Unmarshal(decBody); err != nil {
		t.Fatalf("unmarshal body failed: %v", err)
	}
	if dec.TotalMsgCount != orig.TotalMsgCount {
		t.Errorf("total_msg mismatch: got %d, want %d", dec.TotalMsgCount, orig.TotalMsgCount)
	}
	if dec.OnlineDuration != orig.OnlineDuration {
		t.Errorf("online_duration mismatch: got %d, want %d", dec.OnlineDuration, orig.OnlineDuration)
	}
}

// ============================================================
// JT/T 1078 Round-trip Tests (uses JT808 framing)
// ============================================================

func TestJT1078_RoundTrip_RealtimeAVReq(t *testing.T) {
	codec := jt808.NewCodec()
	orig := &jt1078.RealtimeAVReqMessage{
		SeqNum:     1,
		LogicalCh:  1,
		DataType:   0,
		StreamType: 0,
	}
	header := &types.MessageHeader{
		MsgID: jt1078.MsgIDRealtimeAVReq, Phone: "13800138000", SeqNum: 1,
		Version2019: true, ProtocolVer: 1,
	}
	data, err := codec.Encode(header, orig)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}
	msg, err := codec.Decode(data)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	raw, ok := msg.Body.(*jt808.RawMessage)
	if ok {
		dec := &jt1078.RealtimeAVReqMessage{}
		if err := dec.Unmarshal(raw.Data); err != nil {
			t.Fatalf("unmarshal failed: %v", err)
		}
		if dec.LogicalCh != orig.LogicalCh {
			t.Errorf("logical_ch mismatch: got %d, want %d", dec.LogicalCh, orig.LogicalCh)
		}
	}
}

func TestJT1078_RoundTrip_PTZControl(t *testing.T) {
	codec := jt808.NewCodec()
	orig := &jt1078.PTZControlMessage{
		LogicalCh: 3,
		Direction: 1,
		Speed:     128,
	}
	header := &types.MessageHeader{
		MsgID: jt1078.MsgIDPTZControl, Phone: "13800138000", SeqNum: 1,
		Version2019: true, ProtocolVer: 1,
	}
	data, err := codec.Encode(header, orig)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}
	msg, err := codec.Decode(data)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	raw, ok := msg.Body.(*jt808.RawMessage)
	if ok {
		dec := &jt1078.PTZControlMessage{}
		if err := dec.Unmarshal(raw.Data); err != nil {
			t.Fatalf("unmarshal failed: %v", err)
		}
		if dec.Direction != orig.Direction {
			t.Errorf("direction mismatch: got %d, want %d", dec.Direction, orig.Direction)
		}
		if dec.Speed != orig.Speed {
			t.Errorf("speed mismatch: got %d, want %d", dec.Speed, orig.Speed)
		}
	}
}

func TestJT1078_RoundTrip_PlaybackReq(t *testing.T) {
	codec := jt808.NewCodec()
	orig := &jt1078.PlaybackReqMessage{
		SeqNum:     1,
		LogicalCh:  1,
		StartTime:  "230901000000",
		EndTime:    "230923235959",
		DataType:   0,
		StreamType: 0,
	}
	header := &types.MessageHeader{
		MsgID: jt1078.MsgIDPlaybackReq, Phone: "13800138000", SeqNum: 1,
		Version2019: true, ProtocolVer: 1,
	}
	data, err := codec.Encode(header, orig)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}
	msg, err := codec.Decode(data)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	raw, ok := msg.Body.(*jt808.RawMessage)
	if ok {
		dec := &jt1078.PlaybackReqMessage{}
		if err := dec.Unmarshal(raw.Data); err != nil {
			t.Fatalf("unmarshal failed: %v", err)
		}
		if dec.StartTime != orig.StartTime {
			t.Errorf("start_time mismatch: got %q, want %q", dec.StartTime, orig.StartTime)
		}
		if dec.EndTime != orig.EndTime {
			t.Errorf("end_time mismatch: got %q, want %q", dec.EndTime, orig.EndTime)
		}
	}
}

// ============================================================
// GB/T 32960 Round-trip Tests
// ============================================================

func TestGBT32960_RoundTrip_VehicleLogin(t *testing.T) {
	vin := "LSGAB52A7DF123456"
	loginTime := "230923123456"
	iccid := "12345678901234567890"
	data := gbt32960.EncodeVehicleLogin(vin, loginTime, iccid, 1)
	t.Logf("encoded: %s", hex.EncodeToString(data))

	pkt, err := gbt32960.Decode(data)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if pkt.Header.CmdFlag != gbt32960.MsgTypeVehicleLogin {
		t.Errorf("cmd_flag mismatch: got 0x%02X", pkt.Header.CmdFlag)
	}
	if pkt.Header.VIN != vin {
		t.Errorf("vin mismatch: got %q, want %q", pkt.Header.VIN, vin)
	}
}

func TestGBT32960_RoundTrip_Heartbeat(t *testing.T) {
	vin := "LSGAB52A7DF123456"
	data := gbt32960.EncodeHeartbeat(vin)
	pkt, err := gbt32960.Decode(data)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if pkt.Header.CmdFlag != gbt32960.MsgTypeHeartbeat {
		t.Errorf("cmd_flag mismatch: got 0x%02X", pkt.Header.CmdFlag)
	}
}

func TestGBT32960_RoundTrip_VehicleLogout(t *testing.T) {
	vin := "LSGAB52A7DF123456"
	data := gbt32960.EncodeVehicleLogout(vin, "230923123456", 1)
	pkt, err := gbt32960.Decode(data)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if pkt.Header.CmdFlag != gbt32960.MsgTypeVehicleLogout {
		t.Errorf("cmd_flag mismatch: got 0x%02X", pkt.Header.CmdFlag)
	}
}

func TestGBT32960_RoundTrip_RealtimeInfo(t *testing.T) {
	vin := "LSGAB52A7DF123456"
	vehData := &gbt32960.VehicleData{
		Status: 0x01, ChargeStatus: 0x01, Mode: 0x01, Speed: 600,
		TotalMileage: 100000, Voltage: 35000, Current: 5000, SOC: 80,
		DCDCStatus: 0x01, Shift: 0x02, Resistance: 500,
	}
	posData := &gbt32960.PositionData{
		ChargeStatus: 0x01, Lat: 22543210, Lon: 113945678,
		Altitude: 500, Direction: 180, Speed: 600,
	}
	items := []gbt32960.RealtimeDataItem{
		{Type: gbt32960.DataItemVehicle, Data: gbt32960.EncodeVehicleData(vehData)},
		{Type: gbt32960.DataItemPosition, Data: gbt32960.EncodePositionData(posData)},
	}
	data := gbt32960.EncodeRealtimeInfo(vin, "230923123456", 1, items)
	pkt, err := gbt32960.Decode(data)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if pkt.Header.CmdFlag != gbt32960.MsgTypeRealtimeInfo {
		t.Errorf("cmd_flag mismatch: got 0x%02X", pkt.Header.CmdFlag)
	}
}

func TestGBT32960_RoundTrip_PlatformLogin(t *testing.T) {
	data := gbt32960.EncodePlatformLogin("PLATFORM01", "230923123456", "password123", "encryptkey1234567")
	pkt, err := gbt32960.Decode(data)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if pkt.Header.CmdFlag != gbt32960.MsgTypePlatformLogin {
		t.Errorf("cmd_flag mismatch: got 0x%02X", pkt.Header.CmdFlag)
	}
}

// ============================================================
// JT/T 905 Round-trip Tests (uses JT808 framing)
// ============================================================

func TestJT905_RoundTrip_TaxiOperate(t *testing.T) {
	codec := jt808.NewCodec()
	orig := &jt905.TaxiOperateMessage{
		VehicleColor: 1,
		VehiclePlate: "粤B12345",
		OperateType:  1,
		GetOnTime:    "230923120000",
		GetOffTime:   "230923123000",
		GetOnLat:     22.543210,
		GetOnLon:     113.945678,
		GetOffLat:    22.550000,
		GetOffLon:    113.950000,
		Distance:     5000,
		Amount:       2500,
	}
	header := &types.MessageHeader{
		MsgID: jt905.MsgIDTaxiOperate, Phone: "13800138000", SeqNum: 1,
		Version2019: true, ProtocolVer: 1,
	}
	data, err := codec.Encode(header, orig)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}
	msg, err := codec.Decode(data)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	raw, ok := msg.Body.(*jt808.RawMessage)
	if ok {
		dec := &jt905.TaxiOperateMessage{}
		if err := dec.Unmarshal(raw.Data); err != nil {
			t.Fatalf("unmarshal failed: %v", err)
		}
		if dec.VehiclePlate != orig.VehiclePlate {
			t.Errorf("plate mismatch: got %q, want %q", dec.VehiclePlate, orig.VehiclePlate)
		}
		if dec.Distance != orig.Distance {
			t.Errorf("distance mismatch: got %d, want %d", dec.Distance, orig.Distance)
		}
		if dec.Amount != orig.Amount {
			t.Errorf("amount mismatch: got %d, want %d", dec.Amount, orig.Amount)
		}
	}
}

func TestJT905_RoundTrip_TaxiStatus(t *testing.T) {
	codec := jt808.NewCodec()
	orig := &jt905.TaxiStatusMessage{
		VehicleColor: 1,
		VehiclePlate: "粤B12345",
		Status:       0,
		ChangeTime:   "230923123456",
		Lat:          22.543210,
		Lon:          113.945678,
	}
	header := &types.MessageHeader{
		MsgID: jt905.MsgIDTaxiStatus, Phone: "13800138000", SeqNum: 1,
		Version2019: true, ProtocolVer: 1,
	}
	data, err := codec.Encode(header, orig)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}
	msg, err := codec.Decode(data)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	raw, ok := msg.Body.(*jt808.RawMessage)
	if ok {
		dec := &jt905.TaxiStatusMessage{}
		if err := dec.Unmarshal(raw.Data); err != nil {
			t.Fatalf("unmarshal failed: %v", err)
		}
		if dec.Status != orig.Status {
			t.Errorf("status mismatch: got %d, want %d", dec.Status, orig.Status)
		}
		if dec.Lat != orig.Lat {
			t.Errorf("lat mismatch: got %f, want %f", dec.Lat, orig.Lat)
		}
	}
}

// ============================================================
// JT/T 1045 Round-trip Tests (uses JT808 framing)
// ============================================================

func TestJT1045_RoundTrip_ADASAlarm(t *testing.T) {
	codec := jt808.NewCodec()
	orig := &jt1045.ADASAlarmMessage{
		VehicleColor: 1,
		VehiclePlate: "粤B12345",
		AlarmTime:    "230923123456",
		AlarmType:    0x0101,
		AlarmLevel:   2,
		AlarmParam:   100,
		Lat:          22.543210,
		Lon:          113.945678,
		Altitude:     50,
		Speed:        600,
		Direction:    180,
	}
	header := &types.MessageHeader{
		MsgID: jt1045.MsgIDADASAlarm, Phone: "13800138000", SeqNum: 1,
		Version2019: true, ProtocolVer: 1,
	}
	data, err := codec.Encode(header, orig)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}
	msg, err := codec.Decode(data)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	raw, ok := msg.Body.(*jt808.RawMessage)
	if ok {
		dec := &jt1045.ADASAlarmMessage{}
		if err := dec.Unmarshal(raw.Data); err != nil {
			t.Fatalf("unmarshal failed: %v", err)
		}
		if dec.VehiclePlate != orig.VehiclePlate {
			t.Errorf("plate mismatch: got %q, want %q", dec.VehiclePlate, orig.VehiclePlate)
		}
		if dec.AlarmType != orig.AlarmType {
			t.Errorf("alarm_type mismatch: got 0x%04X, want 0x%04X", dec.AlarmType, orig.AlarmType)
		}
		if dec.Altitude != orig.Altitude {
			t.Errorf("altitude mismatch: got %d, want %d", dec.Altitude, orig.Altitude)
		}
	}
}

func TestJT1045_RoundTrip_DSMAlarm(t *testing.T) {
	codec := jt808.NewCodec()
	orig := &jt1045.DSMAlarmMessage{
		VehicleColor: 1,
		VehiclePlate: "粤B12345",
		AlarmTime:    "230923123456",
		AlarmType:    0x0201,
		AlarmLevel:   1,
		AlarmParam:   50,
		Lat:          22.543210,
		Lon:          113.945678,
		Speed:        600,
		Direction:    180,
	}
	header := &types.MessageHeader{
		MsgID: jt1045.MsgIDDSMAlarm, Phone: "13800138000", SeqNum: 1,
		Version2019: true, ProtocolVer: 1,
	}
	data, err := codec.Encode(header, orig)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}
	msg, err := codec.Decode(data)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	raw, ok := msg.Body.(*jt808.RawMessage)
	if ok {
		dec := &jt1045.DSMAlarmMessage{}
		if err := dec.Unmarshal(raw.Data); err != nil {
			t.Fatalf("unmarshal failed: %v", err)
		}
		if dec.AlarmType != orig.AlarmType {
			t.Errorf("alarm_type mismatch: got 0x%04X, want 0x%04X", dec.AlarmType, orig.AlarmType)
		}
		if dec.Speed != orig.Speed {
			t.Errorf("speed mismatch: got %d, want %d", dec.Speed, orig.Speed)
		}
	}
}

func TestJT1045_RoundTrip_TireAlarm(t *testing.T) {
	codec := jt808.NewCodec()
	orig := &jt1045.TireAlarmMessage{
		VehicleColor: 1,
		VehiclePlate: "粤B12345",
		AlarmTime:    "230923123456",
		TirePosition: 1,
		Pressure:     2500,
		Temp:         35,
		Battery:      80,
		AlarmFlag:    0x01,
	}
	header := &types.MessageHeader{
		MsgID: jt1045.MsgIDTireAlarm, Phone: "13800138000", SeqNum: 1,
		Version2019: true, ProtocolVer: 1,
	}
	data, err := codec.Encode(header, orig)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}
	msg, err := codec.Decode(data)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	raw, ok := msg.Body.(*jt808.RawMessage)
	if ok {
		dec := &jt1045.TireAlarmMessage{}
		if err := dec.Unmarshal(raw.Data); err != nil {
			t.Fatalf("unmarshal failed: %v", err)
		}
		if dec.Pressure != orig.Pressure {
			t.Errorf("pressure mismatch: got %d, want %d", dec.Pressure, orig.Pressure)
		}
		if dec.Temp != orig.Temp {
			t.Errorf("temp mismatch: got %d, want %d", dec.Temp, orig.Temp)
		}
	}
}

// ============================================================
// JT/T 1253 Round-trip Tests (uses JT808 framing)
// ============================================================

func TestJT1253_RoundTrip_EWaybill(t *testing.T) {
	codec := jt808.NewCodec()
	orig := &jt1253.EWaybillMessage{
		VehicleColor: 1,
		VehiclePlate: "粤B12345",
		WaybillID:    "WB20230923001",
		HazardClass:  3,
		HazardName:   "汽油",
		Weight:       10000,
		Origin:       "深圳",
		Destination:  "广州",
		LoadTime:     "230923080000",
		UnloadTime:   "230923120000",
	}
	header := &types.MessageHeader{
		MsgID: jt1253.MsgIDEWaybill, Phone: "13800138000", SeqNum: 1,
		Version2019: true, ProtocolVer: 1,
	}
	data, err := codec.Encode(header, orig)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}
	msg, err := codec.Decode(data)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	raw, ok := msg.Body.(*jt808.RawMessage)
	if ok {
		dec := &jt1253.EWaybillMessage{}
		if err := dec.Unmarshal(raw.Data); err != nil {
			t.Fatalf("unmarshal failed: %v", err)
		}
		if dec.VehiclePlate != orig.VehiclePlate {
			t.Errorf("plate mismatch: got %q, want %q", dec.VehiclePlate, orig.VehiclePlate)
		}
		if dec.WaybillID != orig.WaybillID {
			t.Errorf("waybill_id mismatch: got %q, want %q", dec.WaybillID, orig.WaybillID)
		}
		if dec.HazardClass != orig.HazardClass {
			t.Errorf("hazard_class mismatch: got %d, want %d", dec.HazardClass, orig.HazardClass)
		}
	}
}

func TestJT1253_RoundTrip_HazardAlarm(t *testing.T) {
	codec := jt808.NewCodec()
	orig := &jt1253.HazardAlarmMessage{
		VehicleColor: 1,
		VehiclePlate: "粤B12345",
		AlarmType:    0x0101,
		AlarmLevel:   2,
		AlarmTime:    "230923123456",
		Lat:          22.543210,
		Lon:          113.945678,
		Speed:        600,
		Direction:    180,
	}
	header := &types.MessageHeader{
		MsgID: jt1253.MsgIDHazardAlarm, Phone: "13800138000", SeqNum: 1,
		Version2019: true, ProtocolVer: 1,
	}
	data, err := codec.Encode(header, orig)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}
	msg, err := codec.Decode(data)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	raw, ok := msg.Body.(*jt808.RawMessage)
	if ok {
		dec := &jt1253.HazardAlarmMessage{}
		if err := dec.Unmarshal(raw.Data); err != nil {
			t.Fatalf("unmarshal failed: %v", err)
		}
		if dec.VehiclePlate != orig.VehiclePlate {
			t.Errorf("plate mismatch: got %q, want %q", dec.VehiclePlate, orig.VehiclePlate)
		}
		if dec.AlarmType != orig.AlarmType {
			t.Errorf("alarm_type mismatch: got 0x%04X, want 0x%04X", dec.AlarmType, orig.AlarmType)
		}
	}
}