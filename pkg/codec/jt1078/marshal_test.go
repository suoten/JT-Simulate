package jt1078

import (
	"bytes"
	"testing"
)

// TestMarshalNoSideEffects Marshal 不得修改结构体的时间字段（前导零补齐只作用于局部副本）
func TestMarshalNoSideEffects(t *testing.T) {
	// PlaybackCtrlMessage.PlayTime
	pc := &PlaybackCtrlMessage{
		LogicalCh: 1,
		CtrlCmd:   0,
		PlaySpeed: 0,
		PlayTime:  "230101000000", // 12 字节，需补零到 14
	}
	b1, err := pc.Marshal()
	if err != nil {
		t.Fatalf("marshal playback ctrl failed: %v", err)
	}
	if len(b1) != 10 {
		t.Errorf("playback ctrl marshal length = %d, want 10", len(b1))
	}
	if pc.PlayTime != "230101000000" {
		t.Errorf("Marshal modified PlayTime: got %q, want %q", pc.PlayTime, "230101000000")
	}
	b2, _ := pc.Marshal()
	if !bytes.Equal(b1, b2) {
		t.Error("repeated Marshal produced different output for PlaybackCtrlMessage")
	}

	// PlaybackRespMessage.StartTime / EndTime
	pr := &PlaybackRespMessage{
		RespSeqNum: 1,
		LogicalCh:  1,
		Result:     0,
		StartTime:  "230101000000",
		EndTime:    "230101235959",
	}
	r1, err := pr.Marshal()
	if err != nil {
		t.Fatalf("marshal playback resp failed: %v", err)
	}
	if len(r1) != 18 {
		t.Errorf("playback resp marshal length = %d, want 18", len(r1))
	}
	if pr.StartTime != "230101000000" {
		t.Errorf("Marshal modified StartTime: got %q, want %q", pr.StartTime, "230101000000")
	}
	if pr.EndTime != "230101235959" {
		t.Errorf("Marshal modified EndTime: got %q, want %q", pr.EndTime, "230101235959")
	}
	r2, _ := pr.Marshal()
	if !bytes.Equal(r1, r2) {
		t.Error("repeated Marshal produced different output for PlaybackRespMessage")
	}
}
