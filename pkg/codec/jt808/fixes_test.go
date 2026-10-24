package jt808

import (
	"encoding/binary"
	"math"
	"strings"
	"testing"

	"github.com/suoten/jt-simulate/pkg/types"
)

// ============================== 任务1：0x0200 负坐标 round-trip ==============================

func TestLocationMessageNegativeCoordsRoundTrip(t *testing.T) {
	codec := NewCodec()

	cases := []struct {
		name string
		lat  float64
		lon  float64
	}{
		{"south_west", -33.918861, -118.243683},
		{"south_east", -22.543210, 113.945678},
		{"north_west", 39.904200, -116.407396},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			orig := &LocationMessage{
				AlarmFlag:  0x00000001,
				StatusFlag: 0x00000002,
				Latitude:   tc.lat,
				Longitude:  tc.lon,
				Altitude:   100,
				Speed:      600,
				Direction:  180,
				Time:       "230923123456",
			}
			header := &types.MessageHeader{
				MsgID: MsgIDLocation, Phone: "13800138000", SeqNum: 1,
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
			loc, ok := msg.Body.(*LocationMessage)
			if !ok {
				t.Fatalf("expected *LocationMessage, got %T", msg.Body)
			}
			if loc.Latitude != orig.Latitude {
				t.Errorf("latitude sign lost: got %f, want %f", loc.Latitude, orig.Latitude)
			}
			if loc.Longitude != orig.Longitude {
				t.Errorf("longitude sign lost: got %f, want %f", loc.Longitude, orig.Longitude)
			}
		})
	}
}

func TestLocationMessageMarshalStatusBits(t *testing.T) {
	m := &LocationMessage{
		Latitude:  -1.5,
		Longitude: -2.5,
		Time:      "230923123456",
	}
	data, err := m.Marshal()
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	status := binary.BigEndian.Uint32(data[4:8])
	if status&StatusFlagSouthLatitude == 0 {
		t.Errorf("south latitude status bit (bit2) not set, status=0x%08X", status)
	}
	if status&StatusFlagWestLongitude == 0 {
		t.Errorf("west longitude status bit (bit3) not set, status=0x%08X", status)
	}

	// 正坐标时不应置位
	m2 := &LocationMessage{
		StatusFlag: StatusFlagSouthLatitude | StatusFlagWestLongitude, // 矛盾输入：状态位已置但坐标为正
		Latitude:   1.5,
		Longitude:  2.5,
		Time:       "230923123456",
	}
	data2, err := m2.Marshal()
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	status2 := binary.BigEndian.Uint32(data2[4:8])
	if status2&StatusFlagSouthLatitude != 0 {
		t.Errorf("south latitude status bit should be cleared for positive latitude, status=0x%08X", status2)
	}
	if status2&StatusFlagWestLongitude != 0 {
		t.Errorf("west longitude status bit should be cleared for positive longitude, status=0x%08X", status2)
	}
}

func TestLocationMessageInvalidCoords(t *testing.T) {
	cases := []struct {
		name string
		lat  float64
		lon  float64
	}{
		{"nan_lat", math.NaN(), 113.0},
		{"nan_lon", 22.0, math.NaN()},
		{"lat_overflow", 90.5, 113.0},
		{"lat_underflow", -90.5, 113.0},
		{"lon_overflow", 22.0, 180.5},
		{"lon_underflow", 22.0, -180.5},
		{"inf", math.Inf(1), 113.0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := &LocationMessage{Latitude: tc.lat, Longitude: tc.lon, Time: "230923123456"}
			if _, err := m.Marshal(); err == nil {
				t.Errorf("expected error for lat=%v lon=%v, got nil", tc.lat, tc.lon)
			}
		})
	}
}

// ============================== 任务2：消息体超长 ==============================

func TestEncodeBodyTooLongWithoutPack(t *testing.T) {
	codec := NewCodec()
	header := &types.MessageHeader{
		MsgID: 0x0F0F, Phone: "13800138000", SeqNum: 1,
		Version2019: false,
	}
	body := &RawMessage{ID: 0x0F0F, Data: make([]byte, MaxBodyLen+1)} // 1024 字节
	_, err := codec.Encode(header, body)
	if err == nil {
		t.Fatal("expected error for 1024-byte body without fragmentation, got nil")
	}
	if !strings.Contains(err.Error(), "exceeds") {
		t.Errorf("unexpected error message: %v", err)
	}

	// 恰好 1023 字节：应成功
	bodyOk := &RawMessage{ID: 0x0F0F, Data: make([]byte, MaxBodyLen)}
	if _, err := codec.Encode(header, bodyOk); err != nil {
		t.Errorf("expected success for %d-byte body, got %v", MaxBodyLen, err)
	}

	// 启用分包时不做该限制（由分包逻辑保证）
	packHeader := &types.MessageHeader{
		MsgID: 0x0F0F, Phone: "13800138000", SeqNum: 1,
		Version2019: false, HasPack: true, PackTotal: 2, PackIndex: 1,
	}
	if _, err := codec.Encode(packHeader, body); err != nil {
		t.Errorf("expected success for fragmented oversized body, got %v", err)
	}
}

// ============================== 任务3：Decode 校验 BodyLen ==============================

// patchRaw_frame 解开帧（去分隔符、反转义），调用 patch 修改后重算校验码并重新封装
func rebuildFrame(t *testing.T, frame []byte, patch func(raw []byte) []byte) []byte {
	t.Helper()
	raw, err := Unescape(StripDelimiter(frame))
	if err != nil {
		t.Fatalf("unescape failed: %v", err)
	}
	raw = patch(raw)
	raw[len(raw)-1] = CalcChecksum(raw[:len(raw)-1])
	return WrapWithDelimiter(Escape(raw))
}

func TestDecodeBodyLenExceedsRemaining(t *testing.T) {
	codec := NewCodec()
	orig := &TerminalGeneralRespMessage{RespSeqNum: 1, RespMsgID: 0x8103, Result: 0}
	header := &types.MessageHeader{
		MsgID: MsgIDTerminalGeneralResp, Phone: "13800138000", SeqNum: 1,
		Version2019: false,
	}
	data, err := codec.Encode(header, orig)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	// 把消息体属性中的长度 +1，使其大于实际剩余字节
	tampered := rebuildFrame(t, data, func(raw []byte) []byte {
		attr := binary.BigEndian.Uint16(raw[2:4])
		binary.BigEndian.PutUint16(raw[2:4], attr+1)
		return raw
	})
	_, err = codec.Decode(tampered)
	if err == nil {
		t.Fatal("expected error when declared body length exceeds remaining bytes, got nil")
	}
	if !strings.Contains(err.Error(), "body length") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestDecodeBodyLenSmallerIgnoresTrailing(t *testing.T) {
	codec := NewCodec()
	orig := &LocationMessage{
		AlarmFlag: 0, StatusFlag: 0,
		Latitude: 22.54321, Longitude: 113.945678,
		Altitude: 100, Speed: 600, Direction: 180,
		Time: "230923123456",
	}
	header := &types.MessageHeader{
		MsgID: MsgIDLocation, Phone: "13800138000", SeqNum: 1,
		Version2019: false,
	}
	data, err := codec.Encode(header, orig)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	// 在消息体后（校验码前）追加3字节填充，消息体属性仍声明28字节
	padded := rebuildFrame(t, data, func(raw []byte) []byte {
		return append(raw[:len(raw)-1], 0x00, 0x11, 0x22, raw[len(raw)-1])
	})

	msg, err := codec.Decode(padded)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	loc, ok := msg.Body.(*LocationMessage)
	if !ok {
		t.Fatalf("expected *LocationMessage, got %T", msg.Body)
	}
	if loc.Latitude != orig.Latitude || loc.Longitude != orig.Longitude {
		t.Errorf("coordinates mismatch: got (%f,%f), want (%f,%f)",
			loc.Latitude, loc.Longitude, orig.Latitude, orig.Longitude)
	}
	if len(loc.ExtraData) != 0 {
		t.Errorf("trailing padding should be ignored, got ExtraData=%v", loc.ExtraData)
	}
}

// ============================== 任务4：区域/路线 BCD 与负坐标 ==============================

func TestAreaSetMessagesBCDError(t *testing.T) {
	cases := []struct {
		name  string
		m     interface{ Marshal() ([]byte, error) }
		field string
	}{
		{
			"circular_empty_start",
			&CircularAreaSetMessage{AreaID: 1, Attr: 0, CenterLat: 22.5, CenterLon: 113.9, Radius: 500, StartTime: "", EndTime: "235959"},
			"start time",
		},
		{
			"circular_empty_end",
			&CircularAreaSetMessage{AreaID: 1, Attr: 0, CenterLat: 22.5, CenterLon: 113.9, Radius: 500, StartTime: "000000", EndTime: ""},
			"end time",
		},
		{
			"rect_empty_start",
			&RectAreaSetMessage{AreaID: 1, Attr: 0, UpperLat: 23, UpperLon: 114, LowerLat: 22, LowerLon: 113, StartTime: "", EndTime: "235959"},
			"start time",
		},
		{
			"rect_empty_end",
			&RectAreaSetMessage{AreaID: 1, Attr: 0, UpperLat: 23, UpperLon: 114, LowerLat: 22, LowerLon: 113, StartTime: "000000", EndTime: ""},
			"end time",
		},
		{
			"polygon_empty_start",
			&PolygonAreaSetMessage{AreaID: 1, Attr: 0x02, StartTime: "", EndTime: "235959", Points: []PolygonPoint{{Lat: 22.5, Lon: 113.9}}},
			"start time",
		},
		{
			"polygon_empty_end",
			&PolygonAreaSetMessage{AreaID: 1, Attr: 0x02, StartTime: "000000", EndTime: "", Points: []PolygonPoint{{Lat: 22.5, Lon: 113.9}}},
			"end time",
		},
		{
			"route_empty_start",
			&RouteSetMessage{RouteID: 1, Attr: 0x02, StartTime: "", EndTime: "235959"},
			"start time",
		},
		{
			"route_empty_end",
			&RouteSetMessage{RouteID: 1, Attr: 0x02, StartTime: "000000", EndTime: ""},
			"end time",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := tc.m.Marshal()
			if err == nil {
				t.Fatalf("expected marshal error for empty %s, got nil (data=%X)", tc.field, data)
			}
			if !strings.Contains(err.Error(), tc.field) {
				t.Errorf("error should mention %s: %v", tc.field, err)
			}
		})
	}
}

func TestAreaSetMessagesNegativeCoords(t *testing.T) {
	// 圆形区域：负坐标取绝对值编码
	circular := &CircularAreaSetMessage{
		AreaID: 1, Attr: 0,
		CenterLat: -22.5, CenterLon: -113.9, Radius: 500,
		StartTime: "080000", EndTime: "180000",
	}
	data, err := circular.Marshal()
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	if got := binary.BigEndian.Uint32(data[6:10]); got != uint32(22.5*CoordScaleFactor) {
		t.Errorf("circular lat: got %d, want %d (abs encoded)", got, uint32(22.5*CoordScaleFactor))
	}
	if got := binary.BigEndian.Uint32(data[10:14]); got != uint32(113.9*CoordScaleFactor) {
		t.Errorf("circular lon: got %d, want %d (abs encoded)", got, uint32(113.9*CoordScaleFactor))
	}

	// 矩形区域
	rect := &RectAreaSetMessage{
		AreaID: 1, Attr: 0,
		UpperLat: -23.0, UpperLon: 114.0, LowerLat: -22.0, LowerLon: -113.0,
		StartTime: "080000", EndTime: "180000",
	}
	data, err = rect.Marshal()
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	if got := binary.BigEndian.Uint32(data[6:10]); got != uint32(23.0*CoordScaleFactor) {
		t.Errorf("rect upper lat: got %d, want abs %d", got, uint32(23.0*CoordScaleFactor))
	}
	if got := binary.BigEndian.Uint32(data[18:22]); got != uint32(113.0*CoordScaleFactor) {
		t.Errorf("rect lower lon: got %d, want abs %d", got, uint32(113.0*CoordScaleFactor))
	}

	// 多边形区域
	poly := &PolygonAreaSetMessage{
		AreaID: 1, Attr: 0,
		Points: []PolygonPoint{{Lat: -22.5, Lon: -113.9}, {Lat: 23.5, Lon: 114.9}},
	}
	data, err = poly.Marshal()
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	if got := binary.BigEndian.Uint32(data[8:12]); got != uint32(22.5*CoordScaleFactor) {
		t.Errorf("polygon point0 lat: got %d, want abs %d", got, uint32(22.5*CoordScaleFactor))
	}
	if got := binary.BigEndian.Uint32(data[12:16]); got != uint32(113.9*CoordScaleFactor) {
		t.Errorf("polygon point0 lon: got %d, want abs %d", got, uint32(113.9*CoordScaleFactor))
	}

	// 路线设置
	route := &RouteSetMessage{
		RouteID: 1, Attr: 0,
		Points: []RoutePoint{{PointID: 1, Lat: -22.5, Lon: -113.9, Width: 10, Attr: 0, MaxSpeed: 600, MaxDuration: 60}},
	}
	data, err = route.Marshal()
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	// 布局：id(4)+attr(2)+count(2)+pointID(4) → lat 在偏移 12
	if got := binary.BigEndian.Uint32(data[12:16]); got != uint32(22.5*CoordScaleFactor) {
		t.Errorf("route point lat: got %d, want abs %d", got, uint32(22.5*CoordScaleFactor))
	}
	if got := binary.BigEndian.Uint32(data[16:20]); got != uint32(113.9*CoordScaleFactor) {
		t.Errorf("route point lon: got %d, want abs %d", got, uint32(113.9*CoordScaleFactor))
	}
}

// ============================== 任务5：0x0704 批量位置项长度版本分支 ==============================

func TestGetVersionFeatureBatchItemLenSize(t *testing.T) {
	cases := []struct {
		ver  Version808
		want int
	}{
		{Version808_2011, 1},
		{Version808_2013, 1},
		{Version808_2019, 2},
		{"", 2}, // 零值默认按 2019
	}
	for _, tc := range cases {
		if got := GetVersionFeature(tc.ver).BatchItemLenSize; got != tc.want {
			t.Errorf("GetVersionFeature(%q).BatchItemLenSize = %d, want %d", tc.ver, got, tc.want)
		}
	}
}

func TestLocationBatch2013RoundTrip(t *testing.T) {
	orig := &LocationBatchMessage{
		Type:    1,
		Version: Version808_2013,
		Items: []*LocationMessage{
			{Latitude: -22.54321, Longitude: 113.945678, Altitude: 100, Speed: 600, Direction: 180, Time: "230923123456"},
			{Latitude: 31.234567, Longitude: -118.243683, Altitude: 50, Speed: 300, Direction: 90, Time: "230923130101"},
		},
	}
	data, err := orig.Marshal()
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	// 2013版：type(1)+count(2)+每项 len(1)+item(28)
	if len(data) != 3+2*(1+28) {
		t.Errorf("2013 batch frame length = %d, want %d (1-byte item length field)", len(data), 3+2*29)
	}
	if data[3] != 28 {
		t.Errorf("item0 length field = %d, want 28 (1-byte, 2013)", data[3])
	}

	dec := &LocationBatchMessage{Version: Version808_2013}
	if err := dec.Unmarshal(data); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if len(dec.Items) != len(orig.Items) {
		t.Fatalf("item count = %d, want %d", len(dec.Items), len(orig.Items))
	}
	for i, want := range orig.Items {
		got := dec.Items[i]
		if got.Latitude != want.Latitude || got.Longitude != want.Longitude {
			t.Errorf("item %d coords mismatch: got (%f,%f), want (%f,%f)",
				i, got.Latitude, got.Longitude, want.Latitude, want.Longitude)
		}
	}
}

func TestLocationBatch2019TwoByteLen(t *testing.T) {
	// 2019 版与零值（默认）：长度字段为2字节
	for _, ver := range []Version808{Version808_2019, ""} {
		orig := &LocationBatchMessage{
			Type:    0,
			Version: ver,
			Items:   []*LocationMessage{{Latitude: 22.5, Longitude: 113.9, Time: "230923123456"}},
		}
		data, err := orig.Marshal()
		if err != nil {
			t.Fatalf("marshal (ver=%q) failed: %v", ver, err)
		}
		if len(data) != 3+2+28 {
			t.Errorf("(ver=%q) frame length = %d, want %d (2-byte item length field)", ver, len(data), 3+2+28)
		}
		if got := binary.BigEndian.Uint16(data[3:5]); got != 28 {
			t.Errorf("(ver=%q) item length field = %d, want 28", ver, got)
		}
		dec := &LocationBatchMessage{Version: ver}
		if err := dec.Unmarshal(data); err != nil {
			t.Fatalf("unmarshal (ver=%q) failed: %v", ver, err)
		}
		if len(dec.Items) != 1 || dec.Items[0].Latitude != 22.5 {
			t.Errorf("(ver=%q) round-trip mismatch: %+v", ver, dec.Items)
		}
	}
}

func TestLocationBatch2013OversizeItem(t *testing.T) {
	// 2013 版 1 字节长度字段：单项超过 255 字节应报错
	loc := &LocationMessage{
		Latitude: 22.5, Longitude: 113.9, Time: "230923123456",
		ExtraData: make([]byte, 300),
	}
	m := &LocationBatchMessage{Type: 1, Version: Version808_2013, Items: []*LocationMessage{loc}}
	if _, err := m.Marshal(); err == nil {
		t.Error("expected error for >255-byte item with 1-byte length field (2013), got nil")
	}
}

// ============================== 任务6：CommandRespMessage 一致性 ==============================

func TestCommandRespMarshalCountConsistency(t *testing.T) {
	// Params 为空：ParamCount 编码为 0 且不写参数段
	m := &CommandRespMessage{RespSeqNum: 7, Result: 0, ParamCount: 5}
	data, err := m.Marshal()
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	if len(data) != 4 {
		t.Errorf("empty Params: length = %d, want 4", len(data))
	}
	if data[3] != 0 {
		t.Errorf("empty Params: ParamCount encoded = %d, want 0 (len(Params) is authoritative)", data[3])
	}

	// ParamCount 与 len(Params) 不一致：以 len(Params) 为准
	m2 := &CommandRespMessage{RespSeqNum: 7, Result: 0, ParamCount: 1, Params: []uint32{0x00000001, 0x00000002}}
	data2, err := m2.Marshal()
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	if len(data2) != 4+2*4 {
		t.Errorf("2 params: length = %d, want %d", len(data2), 4+2*4)
	}
	if data2[3] != 2 {
		t.Errorf("ParamCount encoded = %d, want 2 (recomputed from len(Params))", data2[3])
	}

	// round-trip
	dec := &CommandRespMessage{}
	if err := dec.Unmarshal(data2); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if dec.ParamCount != 2 || len(dec.Params) != 2 || dec.Params[0] != 1 || dec.Params[1] != 2 {
		t.Errorf("round-trip mismatch: count=%d params=%v", dec.ParamCount, dec.Params)
	}
}
