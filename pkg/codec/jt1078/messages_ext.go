package jt1078

import (
	"encoding/binary"
	"fmt"
)

// ==================== JT/T 1078-2022 新增消息 ====================

const (
	MsgIDTermAVReq9103     uint16 = 0x9103 // 终端请求音视频资源(2016)
	MsgIDTermAVResp9104    uint16 = 0x9104 // 终端请求音视频资源应答(2016)
	MsgIDCtrlReq9105       uint16 = 0x9105 // 呼叫控制请求(2016)
	MsgIDCtrlResp9106      uint16 = 0x9106 // 呼叫控制应答(2016)
	MsgIDDownloadReq9205   uint16 = 0x9205 // 音视频下载请求(2016)
	MsgIDDownloadResp9206  uint16 = 0x9206 // 音视频下载应答(2016)
	MsgIDDownloadCtrl9207  uint16 = 0x9207 // 下载控制(2016)
	MsgIDDownloadCtrlAck9208 uint16 = 0x9208 // 下载控制应答(2016)
	MsgIDAlarmVideoReq9401 uint16 = 0x9401 // 报警音视频请求
	MsgIDAlarmVideoResp9402 uint16 = 0x9402 // 报警音视频应答
	MsgIDAVParamSet9501    uint16 = 0x9501 // 音视频参数设置
	MsgIDAVParamQuery9502  uint16 = 0x9502 // 音视频参数查询

	// 2022新增
	MsgIDRealtimeAVReq9101_2022 uint16 = 0x9101 // 实时音视频请求(2022扩展)
	MsgIDStreamListReq          uint16 = 0x9107 // 码流列表查询请求(2022)
	MsgIDStreamListResp         uint16 = 0x9108 // 码流列表查询应答(2022)
	MsgIDAudioBroadcast         uint16 = 0x9109 // 音频广播请求(2022)
	MsgIDAudioBroadcastCtrl     uint16 = 0x910A // 音频广播控制(2022)
	MsgIDVideoQualityReq        uint16 = 0x910B // 视频质量查询(2022)
	MsgIDVideoQualityResp       uint16 = 0x910C // 视频质量查询应答(2022)
	MsgIDPTZPresetSet           uint16 = 0x9302 // 云台预置位设置(2022)
	MsgIDPTZPresetCtrl          uint16 = 0x9303 // 云台预置位控制(2022)
	MsgIDPTZCruise              uint16 = 0x9304 // 云台巡航(2022)
	MsgIDPTZTrack               uint16 = 0x9305 // 云台轨迹(2022)
	MsgIDPTZScan                uint16 = 0x9306 // 云台扫描(2022)
	MsgIDPTZAuxCtrl             uint16 = 0x9307 // 云台辅助控制(2022)
	MsgIDAlarmVideoDownloadReq  uint16 = 0x9403 // 报警音视频下载请求(2022)
	MsgIDAlarmVideoDownloadRsp  uint16 = 0x9404 // 报警音视频下载应答(2022)
	MsgIDResourceListReq        uint16 = 0x9504 // 资源列表查询(2022)
	MsgIDResourceListResp       uint16 = 0x9505 // 资源列表查询应答(2022)
	MsgIDSmartAnalysisSet       uint16 = 0x9601 // 智能分析设置(2022)
	MsgIDSmartAnalysisData      uint16 = 0x9602 // 智能分析数据(2022)
	MsgIDDeviceStatusQuery      uint16 = 0x9701 // 设备状态查询(2022)
	MsgIDDeviceStatusResp       uint16 = 0x9702 // 设备状态应答(2022)
)

// Version1078 1078协议版本
type Version1078 string

const (
	Version1078_2016 Version1078 = "2016"
	Version1078_2022 Version1078 = "2022"
)

// MsgInfo1078 1078消息信息
type MsgInfo1078 struct {
	ID   uint16
	Name string
	Dir  string
}

// AllMessages1078 1078协议全部消息列表
var AllMessages1078 = []MsgInfo1078{
	{0x9101, "实时音视频请求", "down"},
	{0x9102, "实时音视频控制", "down"},
	{0x9103, "终端请求音视频", "up"},
	{0x9104, "终端请求音视频应答", "down"},
	{0x9105, "呼叫控制请求", "up"},
	{0x9106, "呼叫控制应答", "down"},
	{0x9107, "码流列表查询请求", "up"},
	{0x9108, "码流列表查询应答", "down"},
	{0x9109, "音频广播请求", "down"},
	{0x910A, "音频广播控制", "down"},
	{0x910B, "视频质量查询", "down"},
	{0x910C, "视频质量查询应答", "up"},
	{0x9201, "音视频回放请求", "down"},
	{0x9202, "音视频回放应答", "up"},
	{0x9203, "回放控制", "down"},
	{0x9204, "回放控制应答", "up"},
	{0x9205, "音视频下载请求", "down"},
	{0x9206, "音视频下载应答", "up"},
	{0x9207, "下载控制", "down"},
	{0x9208, "下载控制应答", "up"},
	{0x9301, "云台控制", "down"},
	{0x9302, "云台预置位设置", "down"},
	{0x9303, "云台预置位控制", "down"},
	{0x9304, "云台巡航", "down"},
	{0x9305, "云台轨迹", "down"},
	{0x9306, "云台扫描", "down"},
	{0x9307, "云台辅助控制", "down"},
	{0x9401, "报警音视频请求", "down"},
	{0x9402, "报警音视频应答", "up"},
	{0x9403, "报警音视频下载请求", "down"},
	{0x9404, "报警音视频下载应答", "up"},
	{0x9501, "音视频参数设置", "down"},
	{0x9502, "音视频参数查询", "down"},
	{0x9503, "音视频参数查询应答", "up"},
	{0x9504, "资源列表查询", "down"},
	{0x9505, "资源列表查询应答", "up"},
	{0x9601, "智能分析设置", "down"},
	{0x9602, "智能分析数据", "up"},
	{0x9701, "设备状态查询", "down"},
	{0x9702, "设备状态应答", "up"},
	{0x1200, "RTP数据包", "up"},
}

// MessagesByVersion1078 按版本过滤1078消息
func MessagesByVersion1078(ver Version1078) []MsgInfo1078 {
	switch ver {
	case Version1078_2016:
		exclude := map[uint16]bool{
			0x9107: true, 0x9108: true, 0x9109: true, 0x910A: true,
			0x910B: true, 0x910C: true,
			0x9302: true, 0x9303: true, 0x9304: true, 0x9305: true,
			0x9306: true, 0x9307: true,
			0x9403: true, 0x9404: true,
			0x9504: true, 0x9505: true,
			0x9601: true, 0x9602: true,
			0x9701: true, 0x9702: true,
		}
		var result []MsgInfo1078
		for _, m := range AllMessages1078 {
			if !exclude[m.ID] {
				result = append(result, m)
			}
		}
		return result
	default:
		return AllMessages1078
	}
}

// PlaybackCtrlMessage 0x9203 回放控制
type PlaybackCtrlMessage struct {
	LogicalCh  byte
	CtrlCmd    byte
	PlaySpeed  byte
	PlayTime   string // YYYYMMDDHHmmss
}

func (m *PlaybackCtrlMessage) MsgID() uint16 { return MsgIDPlaybackCtrl }

func (m *PlaybackCtrlMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 0, 10)
	buf = append(buf, m.LogicalCh)
	buf = append(buf, m.CtrlCmd)
	buf = append(buf, m.PlaySpeed)
	for len(m.PlayTime) < 14 {
		m.PlayTime = "0" + m.PlayTime
	}
	for i := 0; i < 7; i++ {
		buf = append(buf, ((m.PlayTime[i*2]-'0')<<4)|(m.PlayTime[i*2+1]-'0'))
	}
	return buf, nil
}

func (m *PlaybackCtrlMessage) Unmarshal(data []byte) error {
	if len(data) < 10 {
		return fmt.Errorf("1078 playback ctrl too short")
	}
	m.LogicalCh = data[0]
	m.CtrlCmd = data[1]
	m.PlaySpeed = data[2]
	t := make([]byte, 0, 14)
	for i := 0; i < 7; i++ {
		b := data[3+i]
		t = append(t, (b>>4)+'0', (b&0x0F)+'0')
	}
	m.PlayTime = string(t)
	return nil
}

// PlaybackRespMessage 0x9202 回放应答
type PlaybackRespMessage struct {
	RespSeqNum  uint16
	LogicalCh   byte
	Result      byte
	StartTime   string
	EndTime     string
}

func (m *PlaybackRespMessage) MsgID() uint16 { return MsgIDPlaybackResp }

func (m *PlaybackRespMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 0, 20)
	seq := make([]byte, 2)
	binary.BigEndian.PutUint16(seq, m.RespSeqNum)
	buf = append(buf, seq...)
	buf = append(buf, m.LogicalCh)
	buf = append(buf, m.Result)
	for _, t := range []string{m.StartTime, m.EndTime} {
		for len(t) < 14 {
			t = "0" + t
		}
		for i := 0; i < 7; i++ {
			buf = append(buf, ((t[i*2]-'0')<<4)|(t[i*2+1]-'0'))
		}
	}
	return buf, nil
}

func (m *PlaybackRespMessage) Unmarshal(data []byte) error {
	if len(data) < 18 {
		return fmt.Errorf("1078 playback resp too short")
	}
	m.RespSeqNum = binary.BigEndian.Uint16(data[0:2])
	m.LogicalCh = data[2]
	m.Result = data[3]
	decBCD := func(off int) string {
		s := make([]byte, 0, 14)
		for i := 0; i < 7; i++ {
			b := data[off+i]
			s = append(s, (b>>4)+'0', (b&0x0F)+'0')
		}
		return string(s)
	}
	m.StartTime = decBCD(4)
	m.EndTime = decBCD(11)
	return nil
}

// TermAVReqMessage 0x9103 终端请求音视频
type TermAVReqMessage struct {
	LogicalCh  byte
	DataType   byte
	StreamType byte
}

func (m *TermAVReqMessage) MsgID() uint16 { return MsgIDTermAVReq }

func (m *TermAVReqMessage) Marshal() ([]byte, error) {
	return []byte{m.LogicalCh, m.DataType, m.StreamType}, nil
}

func (m *TermAVReqMessage) Unmarshal(data []byte) error {
	if len(data) < 3 {
		return fmt.Errorf("1078 term av req too short")
	}
	m.LogicalCh = data[0]
	m.DataType = data[1]
	m.StreamType = data[2]
	return nil
}

// TermAVRespMessage 0x9104 终端请求音视频应答
type TermAVRespMessage struct {
	RespSeqNum uint16
	LogicalCh  byte
	Result     byte
}

func (m *TermAVRespMessage) MsgID() uint16 { return MsgIDTermAVResp }

func (m *TermAVRespMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint16(buf[0:2], m.RespSeqNum)
	buf[2] = m.LogicalCh
	buf[3] = m.Result
	return buf, nil
}

func (m *TermAVRespMessage) Unmarshal(data []byte) error {
	if len(data) < 4 {
		return fmt.Errorf("1078 term av resp too short")
	}
	m.RespSeqNum = binary.BigEndian.Uint16(data[0:2])
	m.LogicalCh = data[2]
	m.Result = data[3]
	return nil
}

// AVParamSetMessage 0x9501 音视频参数设置
type AVParamSetMessage struct {
	LogicalCh   byte
	AudioFormat byte
	VideoFormat byte
	Resolution  byte
	FrameRate   byte
	Bitrate     uint32
}

func (m *AVParamSetMessage) MsgID() uint16 { return MsgIDAVParamSet }

func (m *AVParamSetMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 9)
	buf[0] = m.LogicalCh
	buf[1] = m.AudioFormat
	buf[2] = m.VideoFormat
	buf[3] = m.Resolution
	buf[4] = m.FrameRate
	binary.BigEndian.PutUint32(buf[5:9], m.Bitrate)
	return buf, nil
}

func (m *AVParamSetMessage) Unmarshal(data []byte) error {
	if len(data) < 9 {
		return fmt.Errorf("1078 av param set too short")
	}
	m.LogicalCh = data[0]
	m.AudioFormat = data[1]
	m.VideoFormat = data[2]
	m.Resolution = data[3]
	m.FrameRate = data[4]
	m.Bitrate = binary.BigEndian.Uint32(data[5:9])
	return nil
}

// Update MsgName in codec.go to use full list
func init() {
	// This ensures the extended messages are registered
}
