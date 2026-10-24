package jt809

import (
	"bytes"
	"testing"
)

// TestEncodeHeaderVersionPadding 验证 Version 不足 3 字节时保留已有字节、后位补零，
// 不再整体丢弃换成全零；且 EncodeHeader 不修改结构体字段。
func TestEncodeHeaderVersionPadding(t *testing.T) {
	c := NewCodec()
	cases := []struct {
		name    string
		version []byte
		flag    uint8
		want    []byte
	}{
		{"1字节后位补零", []byte{0x01}, 0, []byte{0x01, 0x00, 0x00}},
		{"2字节后位补零", []byte{0x01, 0x02}, 1, []byte{0x01, 0x02, 0x00}},
		{"3字节原样保留", []byte{0x02, 0x00, 0x00}, 1, []byte{0x02, 0x00, 0x00}},
		{"空+2019用默认版本", nil, 1, []byte{0x01, 0x00, 0x00}},
		{"空+2011全零", nil, 0, []byte{0x00, 0x00, 0x00}},
		{"超长截断到3字节", []byte{0x01, 0x02, 0x03, 0x04}, 0, []byte{0x01, 0x02, 0x03}},
	}
	for _, tc := range cases {
		h := &Header{
			MsgSN:           1,
			MsgID:           MsgIDUpLinktestReq,
			MsgGNSSCENTERID: 1001,
			VersionFlag:     tc.flag,
			Version:         tc.version,
		}
		buf, err := c.EncodeHeader(h)
		if err != nil {
			t.Fatalf("%s: EncodeHeader failed: %v", tc.name, err)
		}
		if !bytes.Equal(buf[14:17], tc.want) {
			t.Errorf("%s: version bytes = % X, want % X", tc.name, buf[14:17], tc.want)
		}
		// 结构体字段不得被修改
		if !bytes.Equal(h.Version, tc.version) {
			t.Errorf("%s: EncodeHeader modified h.Version: got % X, want % X", tc.name, h.Version, tc.version)
		}
	}
}

// TestDecode2019Fallback 覆盖 2019 版兜底检测路径：
// Version[0] != 0x01 的 2019 版报文，借助消息头中的 UTC 时间戳字段被正确识别。
func TestDecode2019Fallback(t *testing.T) {
	c := NewCodec()
	body := []byte{0x01, 0x02, 0x03, 0x04}
	h := &Header{
		MsgSN:           5,
		MsgID:           MsgIDLocationMsg,
		MsgGNSSCENTERID: 1001,
		VersionFlag:     1,
		Version:         []byte{0x02, 0x00, 0x00}, // Version[0] != 0x01，主检测失效
		UTCTime:         1700000000000,            // 合理的毫秒级 Unix 时间戳
	}
	frame, err := c.Encode(h, body)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	h2, got, err := c.Decode(frame)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if h2.VersionFlag != 1 {
		t.Fatalf("VersionFlag = %d, want 1 (2019 fallback path)", h2.VersionFlag)
	}
	if h2.UTCTime != h.UTCTime {
		t.Errorf("UTCTime = %d, want %d", h2.UTCTime, h.UTCTime)
	}
	if !bytes.Equal(h2.Version, []byte{0x02, 0x00, 0x00}) {
		t.Errorf("decoded Version = % X, want 02 00 00", h2.Version)
	}
	if !bytes.Equal(got, body) {
		t.Errorf("body round-trip mismatch:\n got: % X\nwant: % X", got, body)
	}
}

// TestDecode2019FallbackSecondsTimestamp 秒级时间戳同样应触发 2019 兜底
func TestDecode2019FallbackSecondsTimestamp(t *testing.T) {
	c := NewCodec()
	h := &Header{
		MsgSN:           6,
		MsgID:           MsgIDUpLinktestReq,
		MsgGNSSCENTERID: 1001,
		VersionFlag:     1,
		Version:         []byte{0x02, 0x00, 0x00},
		UTCTime:         1700000000, // 秒级
	}
	frame, err := c.Encode(h, nil)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}
	h2, _, err := c.Decode(frame)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if h2.VersionFlag != 1 {
		t.Errorf("VersionFlag = %d, want 1 (2019 fallback, seconds timestamp)", h2.VersionFlag)
	}
}

// TestDecode2011LongBodyNotMisdetectedAs2019 对照组：
// 长消息体（头+体 >= 30 字节）的 2011 版报文不应被兜底逻辑误判为 2019 版。
func TestDecode2011LongBodyNotMisdetectedAs2019(t *testing.T) {
	c := NewCodec()
	loc := &LocationMessage{
		VehicleColor: 2,
		VehiclePlate: "粤B12345",
		AlarmFlag:    1,
		Lat:          22.5,
		Lon:          113.25,
		Speed:        300,
		Direction:    90,
		Time:         "230101120000",
	}
	bodyBytes, err := loc.Marshal()
	if err != nil {
		t.Fatalf("marshal location failed: %v", err)
	}
	if len(bodyBytes) < 8 {
		t.Fatalf("test expects body >= 8 bytes, got %d", len(bodyBytes))
	}
	h := &Header{MsgSN: 5, MsgID: MsgIDLocationMsg, MsgGNSSCENTERID: 1001, VersionFlag: 0}
	frame, err := c.Encode(h, bodyBytes)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}
	h2, got, err := c.Decode(frame)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if h2.VersionFlag != 0 {
		t.Errorf("2011 frame with long body misdetected as 2019 (VersionFlag = %d)", h2.VersionFlag)
	}
	if !bytes.Equal(got, bodyBytes) {
		t.Errorf("body mismatch:\n got: % X\nwant: % X", got, bodyBytes)
	}
}
