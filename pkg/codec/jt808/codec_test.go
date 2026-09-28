package jt808

import (
	"testing"

	"github.com/suoten/jt-simulate/pkg/types"
)

func TestMsgName(t *testing.T) {
	tests := []struct {
		msgID uint16
		want  string
	}{
		{0x0100, "终端注册"},
		{0x0102, "终端鉴权"},
		{0x0200, "位置上报"},
		{0x0002, "终端心跳"},
		{0x8001, "平台通用应答"},
		{0x8100, "终端注册应答"},
	}

	for _, tt := range tests {
		got := MsgName(tt.msgID)
		if got != tt.want {
			t.Errorf("MsgName(0x%04x) = %q, want %q", tt.msgID, got, tt.want)
		}
	}
}

func TestEncodeHeader(t *testing.T) {
	codec := NewCodec()

	header := &types.MessageHeader{
		MsgID:  0x0002, // 心跳 — 无消息体
		Phone:  "013800001111",
		SeqNum: 1,
	}

	headerBytes, err := codec.EncodeHeader(header)
	if err != nil {
		t.Fatalf("EncodeHeader failed: %v", err)
	}

	if len(headerBytes) < 10 {
		t.Fatalf("header too short: %d", len(headerBytes))
	}
}

func TestVerifyChecksum(t *testing.T) {
	codec := NewCodec()

	// 手动构造一个帧（不带转义）
	data := []byte{0x00, 0x02, 0x00, 0x00, 0x01, 0x38, 0x00, 0x00, 0x11, 0x11, 0x00, 0x01}
	checksum := CalcChecksum(data)
	full := append(data, checksum)

	if !codec.VerifyChecksum(full) {
		t.Error("VerifyChecksum should return true for valid frame")
	}

	// 篡改一个字节
	tampered := make([]byte, len(full))
	copy(tampered, full)
	tampered[3] ^= 0xff
	if codec.VerifyChecksum(tampered) {
		t.Error("VerifyChecksum should return false for tampered frame")
	}
}

func TestCalcChecksum(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03, 0x04}
	cs := CalcChecksum(data)
	expected := byte(0x01 ^ 0x02 ^ 0x03 ^ 0x04)
	if cs != expected {
		t.Errorf("expected 0x%02x, got 0x%02x", expected, cs)
	}
}

func TestSplitByDelimiterEmpty(t *testing.T) {
	frames := SplitByDelimiter([]byte{})
	if len(frames) != 0 {
		t.Errorf("expected 0 frames, got %d", len(frames))
	}
}
