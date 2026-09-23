package jt1078

import (
	"encoding/binary"
	"fmt"
)

// 消息ID常量
const (
	MsgIDRealtimeAVReq   uint16 = 0x9101 // 实时音视频请求
	MsgIDRealtimeAVCtrl  uint16 = 0x9102 // 实时音视频控制
	MsgIDTermAVReq       uint16 = 0x9103 // 终端请求音视频
	MsgIDTermAVResp      uint16 = 0x9104 // 终端请求音视频应答
	MsgIDCtrlReq         uint16 = 0x9105 // 呼叫控制请求
	MsgIDCtrlResp        uint16 = 0x9106 // 呼叫控制应答
	MsgIDPlaybackReq     uint16 = 0x9201 // 音视频回放请求
	MsgIDPlaybackResp    uint16 = 0x9202 // 音视频回放应答
	MsgIDPlaybackCtrl    uint16 = 0x9203 // 回放控制
	MsgIDPlaybackCtrlAck uint16 = 0x9204 // 回放控制应答
	MsgIDDownloadReq     uint16 = 0x9205 // 音视频下载请求
	MsgIDDownloadResp    uint16 = 0x9206 // 音视频下载应答
	MsgIDRTPData         uint16 = 0x1200 // RTP数据包
	MsgIDAlarmVideoReq   uint16 = 0x9401 // 报警视频请求
	MsgIDAlarmVideoResp  uint16 = 0x9402 // 报警视频应答
	MsgIDPTZControl      uint16 = 0x9301 // 云台控制
	MsgIDAVParamSet      uint16 = 0x9501 // 音视频参数设置
	MsgIDAVParamQuery    uint16 = 0x9502 // 音视频参数查询
	MsgIDAVParamResp     uint16 = 0x9503 // 音视频参数查询应答
)

// RTPPacketType RTP包类型
const (
	RTPTypeVideo = 0 // 视频
	RTPTypeAudio = 1 // 音频
)

// RTPHeader RTP包头
type RTPHeader struct {
	Version     uint8
	Padding     bool
	Extension   bool
	CC          uint8
	Marker      bool
	PayloadType uint8
	SeqNum      uint16
	Timestamp   uint32
	SSRC        uint32
}

// RealtimeAVReqMessage 0x9101 实时音视频请求
type RealtimeAVReqMessage struct {
	SeqNum      uint16
	LogicalCh   byte
	DataType    byte
	StreamType  byte
}

func (m *RealtimeAVReqMessage) MsgID() uint16 { return MsgIDRealtimeAVReq }

func (m *RealtimeAVReqMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 5)
	binary.BigEndian.PutUint16(buf[0:2], m.SeqNum)
	buf[2] = m.LogicalCh
	buf[3] = m.DataType
	buf[4] = m.StreamType
	return buf, nil
}

func (m *RealtimeAVReqMessage) Unmarshal(data []byte) error {
	if len(data) < 5 {
		return fmt.Errorf("1078 realtime av req too short")
	}
	m.SeqNum = binary.BigEndian.Uint16(data[0:2])
	m.LogicalCh = data[2]
	m.DataType = data[3]
	m.StreamType = data[4]
	return nil
}

// RealtimeAVCtrlMessage 0x9102 实时音视频控制
// JT/T 1078-2016: 逻辑通道号(1) + 控制指令(1) + 切换类型(1) = 3字节
type RealtimeAVCtrlMessage struct {
	LogicalCh  byte // 逻辑通道号
	CtrlCmd    byte // 控制指令：0=关闭 1=切换码流
	SwitchType byte // 切换类型（CtrlCmd=1时有效）：0=主码流 1=子码流
}

func (m *RealtimeAVCtrlMessage) MsgID() uint16 { return MsgIDRealtimeAVCtrl }

func (m *RealtimeAVCtrlMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 3)
	buf[0] = m.LogicalCh
	buf[1] = m.CtrlCmd
	buf[2] = m.SwitchType
	return buf, nil
}

func (m *RealtimeAVCtrlMessage) Unmarshal(data []byte) error {
	if len(data) < 3 {
		return fmt.Errorf("1078 realtime av ctrl too short")
	}
	m.LogicalCh = data[0]
	m.CtrlCmd = data[1]
	m.SwitchType = data[2]
	return nil
}

// PlaybackReqMessage 0x9201 回放请求
type PlaybackReqMessage struct {
	SeqNum     uint16
	LogicalCh  byte
	StartTime  string // YYMMDDHHmmss
	EndTime    string
	DataType   byte
	StreamType byte
}

func (m *PlaybackReqMessage) MsgID() uint16 { return MsgIDPlaybackReq }

func (m *PlaybackReqMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 0, 20)
	seq := make([]byte, 2)
	binary.BigEndian.PutUint16(seq, m.SeqNum)
	buf = append(buf, seq...)
	buf = append(buf, m.LogicalCh)

	// 时间BCD 6字节 each
	for _, timeStr := range []string{m.StartTime, m.EndTime} {
		for len(timeStr) < 12 {
			timeStr = "0" + timeStr
		}
		for i := 0; i < 6; i++ {
			high := timeStr[i*2] - '0'
			low := timeStr[i*2+1] - '0'
			buf = append(buf, (high<<4)|low)
		}
	}

	buf = append(buf, m.DataType)
	buf = append(buf, m.StreamType)
	return buf, nil
}

func (m *PlaybackReqMessage) Unmarshal(data []byte) error {
	if len(data) < 17 {
		return fmt.Errorf("1078 playback req too short")
	}
	m.SeqNum = binary.BigEndian.Uint16(data[0:2])
	m.LogicalCh = data[2]
	decodeBCD := func(offset int) string {
		s := make([]byte, 0, 12)
		for i := 0; i < 6; i++ {
			b := data[offset+i]
			s = append(s, (b>>4)+'0', (b&0x0F)+'0')
		}
		return string(s)
	}
	m.StartTime = decodeBCD(3)
	m.EndTime = decodeBCD(9)
	m.DataType = data[15]
	m.StreamType = data[16]
	return nil
}

// PTZControlMessage 0x9301 云台控制
// JT/T 1078-2016: 逻辑通道号(1) + 方向(1) + 速度(1) = 3字节
type PTZControlMessage struct {
	LogicalCh  byte // 逻辑通道号
	Direction  byte // 方向控制：0=停止 1=上 2=下 3=左 4=右 5=左上 6=右上 7=左下 8=右下 9=变倍缩小 10=变倍放大 11=焦距变小 12=焦距变大 13=光圈变小 14=光圈变大
	Speed      byte // 速度（0-255）
}

func (m *PTZControlMessage) MsgID() uint16 { return MsgIDPTZControl }

func (m *PTZControlMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 3)
	buf[0] = m.LogicalCh
	buf[1] = m.Direction
	buf[2] = m.Speed
	return buf, nil
}

func (m *PTZControlMessage) Unmarshal(data []byte) error {
	if len(data) < 3 {
		return fmt.Errorf("1078 ptz too short")
	}
	m.LogicalCh = data[0]
	m.Direction = data[1]
	m.Speed = data[2]
	return nil
}

// AVParamMessage 0x9503 音视频参数
type AVParamMessage struct {
	SeqNum          uint16
	LogicalCh       byte
	AudioFormat     byte
	AudioChan       byte
	AudioSample     byte
	AudioBit        byte
	VideoFormat     byte
	Resolution      byte
	GPSup           byte
	GPSinterval     byte
}

func (m *AVParamMessage) MsgID() uint16 { return MsgIDAVParamResp }

func (m *AVParamMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 11)
	binary.BigEndian.PutUint16(buf[0:2], m.SeqNum)
	buf[2] = m.LogicalCh
	buf[3] = m.AudioFormat
	buf[4] = m.AudioChan
	buf[5] = m.AudioSample
	buf[6] = m.AudioBit
	buf[7] = m.VideoFormat
	buf[8] = m.Resolution
	buf[9] = m.GPSup
	buf[10] = m.GPSinterval
	return buf, nil
}

func (m *AVParamMessage) Unmarshal(data []byte) error {
	if len(data) < 11 {
		return fmt.Errorf("1078 av param too short")
	}
	m.SeqNum = binary.BigEndian.Uint16(data[0:2])
	m.LogicalCh = data[2]
	m.AudioFormat = data[3]
	m.AudioChan = data[4]
	m.AudioSample = data[5]
	m.AudioBit = data[6]
	m.VideoFormat = data[7]
	m.Resolution = data[8]
	m.GPSup = data[9]
	m.GPSinterval = data[10]
	return nil
}

// RTP RTP包（0x1200 在808协议中作为消息体传输）
type RTPPacket struct {
	Header  RTPHeader
	Payload []byte
}

// EncodeRTP 编码RTP包
func EncodeRTP(pkt *RTPPacket) []byte {
	buf := make([]byte, 12+len(pkt.Payload))
	buf[0] = (pkt.Header.Version << 6) | (boolToByte(pkt.Header.Padding) << 5) | (boolToByte(pkt.Header.Extension) << 4) | (pkt.Header.CC & 0x0F)
	buf[1] = (boolToByte(pkt.Header.Marker) << 7) | (pkt.Header.PayloadType & 0x7F)
	binary.BigEndian.PutUint16(buf[2:4], pkt.Header.SeqNum)
	binary.BigEndian.PutUint32(buf[4:8], pkt.Header.Timestamp)
	binary.BigEndian.PutUint32(buf[8:12], pkt.Header.SSRC)
	copy(buf[12:], pkt.Payload)
	return buf
}

// DecodeRTP 解码RTP包
func DecodeRTP(data []byte) (*RTPPacket, error) {
	if len(data) < 12 {
		return nil, fmt.Errorf("RTP packet too short: %d", len(data))
	}
	pkt := &RTPPacket{
		Header: RTPHeader{
			Version:     (data[0] >> 6) & 0x03,
			Padding:     data[0]&0x20 != 0,
			Extension:   data[0]&0x10 != 0,
			CC:          data[0] & 0x0F,
			Marker:      data[1]&0x80 != 0,
			PayloadType: data[1] & 0x7F,
			SeqNum:      binary.BigEndian.Uint16(data[2:4]),
			Timestamp:   binary.BigEndian.Uint32(data[4:8]),
			SSRC:        binary.BigEndian.Uint32(data[8:12]),
		},
	}
	pkt.Payload = make([]byte, len(data)-12)
	copy(pkt.Payload, data[12:])
	return pkt, nil
}

// MsgName 消息名称
func MsgName(msgID uint16) string {
	for _, m := range AllMessages1078 {
		if m.ID == msgID {
			return m.Name
		}
	}
	return fmt.Sprintf("未知消息(0x%04X)", msgID)
}

func boolToByte(b bool) byte {
	if b {
		return 1
	}
	return 0
}
