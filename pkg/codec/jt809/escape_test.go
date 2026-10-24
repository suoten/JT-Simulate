package jt809

import (
	"bytes"
	"testing"
)

// escapeBytes 按 JT/T 809 标准转义表转义（测试辅助，独立于 Encode 内实现）
// 0x5B → 5A 01；0x5A → 5A 02；0x5D → 5E 01；0x5E → 5E 02
func escapeBytes(raw []byte) []byte {
	out := make([]byte, 0, len(raw)*2)
	for _, b := range raw {
		switch b {
		case 0x5B:
			out = append(out, 0x5A, 0x01)
		case 0x5A:
			out = append(out, 0x5A, 0x02)
		case 0x5D:
			out = append(out, 0x5E, 0x01)
		case 0x5E:
			out = append(out, 0x5E, 0x02)
		default:
			out = append(out, b)
		}
	}
	return out
}

// TestEscapeTableEncode 验证编码转义规则：
// 1. 帧首 0x5B / 帧尾 0x5D 本身不转义；
// 2. 中间部分（含 CRC）按标准表转义，产生 5A 01 / 5A 02 / 5E 01 / 5E 02；
// 3. CRC 对转义前的原始数据计算。
func TestEscapeTableEncode(t *testing.T) {
	c := NewCodec()
	body := []byte{0x5A, 0x5B, 0x5C, 0x5D, 0x5E, 0x00, 0xFF}
	h := &Header{
		MsgSN:           7,
		MsgID:           MsgIDUpLinktestReq,
		MsgGNSSCENTERID: 1001,
		VersionFlag:     1,
		UTCTime:         1700000000000, // 固定时间戳，保证期望帧可精确重建
	}
	frame, err := c.Encode(h, body)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	// 帧首 0x5B，帧尾 0x5D
	if frame[0] != BeginFlag {
		t.Errorf("frame start = 0x%02X, want 0x5B", frame[0])
	}
	if frame[len(frame)-1] != EndFlag {
		t.Errorf("frame end = 0x%02X, want 0x5D", frame[len(frame)-1])
	}

	// 用 Encode 结束后的头部（MsgLength/UTCTime 已填好）重建期望帧：
	// 头 + 体 → 对未转义数据计算 CRC → 整体转义 → 加首尾标记
	headerBytes, err := c.EncodeHeader(h)
	if err != nil {
		t.Fatalf("re-encode header failed: %v", err)
	}
	raw := make([]byte, 0, len(headerBytes)+len(body)+checksumLen)
	raw = append(raw, headerBytes...)
	raw = append(raw, body...)
	crc := calcCRC16(raw)
	raw = append(raw, byte(crc>>8), byte(crc&0xFF))

	want := append([]byte{BeginFlag}, escapeBytes(raw)...)
	want = append(want, EndFlag)

	if !bytes.Equal(frame, want) {
		t.Fatalf("escaped frame mismatch:\n got: % X\nwant: % X", frame, want)
	}

	// 转义后的帧内必须包含四个转义序列（消息体保证出现全部四种特殊字节）
	inner := frame[1 : len(frame)-1]
	for _, seq := range [][]byte{{0x5A, 0x01}, {0x5A, 0x02}, {0x5E, 0x01}, {0x5E, 0x02}} {
		if !bytes.Contains(inner, seq) {
			t.Errorf("escaped frame missing sequence % X", seq)
		}
	}
	// 转义后帧内不得出现裸的 0x5B / 0x5D（否则帧边界会被破坏）
	if bytes.Contains(inner, []byte{0x5B}) || bytes.Contains(inner, []byte{0x5D}) {
		t.Errorf("raw 0x5B/0x5D leaked inside escaped frame: % X", inner)
	}
}

// TestEscapeRoundTripSpecialBytes 含 0x5A/0x5B/0x5D/0x5E 的报文 round-trip 后字节完全一致
func TestEscapeRoundTripSpecialBytes(t *testing.T) {
	c := NewCodec()
	bodies := [][]byte{
		{0x5A},
		{0x5B},
		{0x5D},
		{0x5E},
		{0x5A, 0x5B, 0x5C, 0x5D, 0x5E},
		{0x01, 0x02}, // 转义前缀字节出现在正文中
		{0x5A, 0x01, 0x5A, 0x02, 0x5E, 0x01, 0x5E, 0x02},
		{0x5B, 0x5B, 0x5D, 0x5D, 0x5A, 0x5A, 0x5E, 0x5E},
		bytes.Repeat([]byte{0x5A, 0x5B, 0x5D, 0x5E}, 32),
	}
	for _, flag := range []uint8{0, 1} { // 2011 与 2019 两种头
		for i, body := range bodies {
			h := &Header{
				MsgSN:           1,
				MsgID:           MsgIDLocationMsg,
				MsgGNSSCENTERID: 1001,
				VersionFlag:     flag,
				UTCTime:         1700000000000,
			}
			frame, err := c.Encode(h, body)
			if err != nil {
				t.Fatalf("encode body[%d] flag=%d failed: %v", i, flag, err)
			}
			h2, got, err := c.Decode(frame)
			if err != nil {
				t.Fatalf("decode body[%d] flag=%d failed: %v", i, flag, err)
			}
			if h2.VersionFlag != flag {
				t.Errorf("body[%d] flag=%d: decoded VersionFlag = %d, want %d", i, flag, h2.VersionFlag, flag)
			}
			if !bytes.Equal(got, body) {
				t.Errorf("body[%d] flag=%d round-trip mismatch:\n got: % X\nwant: % X", i, flag, got, body)
			}
		}
	}
}

// TestDecodeHandcraftedEscapedFrame 用独立构造的转义帧验证解码还原：
// 5A 01 → 5B；5A 02 → 5A；5E 01 → 5D；5E 02 → 5E
func TestDecodeHandcraftedEscapedFrame(t *testing.T) {
	c := NewCodec()
	body := []byte{0x5B, 0x5A, 0x5D, 0x5E}
	h := &Header{
		MsgSN:           9,
		MsgID:           MsgIDLocationMsg,
		MsgGNSSCENTERID: 1001,
		VersionFlag:     0,
		Version:         []byte{0x00, 0x00, 0x00},
		MsgLength:       uint32(HeaderLen2011 + len(body) + checksumLen + 1),
	}
	headerBytes, err := c.EncodeHeader(h)
	if err != nil {
		t.Fatalf("encode header failed: %v", err)
	}
	raw := append(headerBytes, body...)
	crc := calcCRC16(raw)
	raw = append(raw, byte(crc>>8), byte(crc&0xFF))

	frame := append([]byte{BeginFlag}, escapeBytes(raw)...)
	frame = append(frame, EndFlag)

	h2, got, err := c.Decode(frame)
	if err != nil {
		t.Fatalf("decode handcrafted frame failed: %v", err)
	}
	if h2.VersionFlag != 0 {
		t.Errorf("decoded VersionFlag = %d, want 0 (2011)", h2.VersionFlag)
	}
	if h2.MsgID != MsgIDLocationMsg {
		t.Errorf("decoded MsgID = 0x%04X, want 0x%04X", h2.MsgID, MsgIDLocationMsg)
	}
	if !bytes.Equal(got, body) {
		t.Errorf("decoded body mismatch:\n got: % X\nwant: % X", got, body)
	}
}
