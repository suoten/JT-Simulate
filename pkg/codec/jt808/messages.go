package jt808

import (
	"encoding/binary"
	"fmt"
	"math"
	"strings"

	"golang.org/x/text/encoding/simplifiedchinese"
)

// RawMessage 原始消息（未识别的消息ID）
type RawMessage struct {
	ID   uint16
	Data []byte
}

func (m *RawMessage) MsgID() uint16         { return m.ID }
func (m *RawMessage) Marshal() ([]byte, error) { return m.Data, nil }
func (m *RawMessage) Unmarshal(data []byte) error {
	m.Data = make([]byte, len(data))
	copy(m.Data, data)
	return nil
}

// ============================== 0x0001 终端通用应答 ==============================

type TerminalGeneralRespMessage struct {
	RespSeqNum   uint16
	RespMsgID    uint16
	Result       byte
}

func (m *TerminalGeneralRespMessage) MsgID() uint16 { return MsgIDTerminalGeneralResp }

func (m *TerminalGeneralRespMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 5)
	binary.BigEndian.PutUint16(buf[0:2], m.RespSeqNum)
	binary.BigEndian.PutUint16(buf[2:4], m.RespMsgID)
	buf[4] = m.Result
	return buf, nil
}

func (m *TerminalGeneralRespMessage) Unmarshal(data []byte) error {
	if len(data) < 5 {
		return ErrDataTooShort
	}
	m.RespSeqNum = binary.BigEndian.Uint16(data[0:2])
	m.RespMsgID = binary.BigEndian.Uint16(data[2:4])
	m.Result = data[4]
	return nil
}

// ============================== 0x8001 平台通用应答 ==============================

type PlatformGeneralRespMessage struct {
	RespSeqNum uint16
	RespMsgID  uint16
	Result     byte
}

func (m *PlatformGeneralRespMessage) MsgID() uint16 { return 0x8001 }

func (m *PlatformGeneralRespMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 5)
	binary.BigEndian.PutUint16(buf[0:2], m.RespSeqNum)
	binary.BigEndian.PutUint16(buf[2:4], m.RespMsgID)
	buf[4] = m.Result
	return buf, nil
}

func (m *PlatformGeneralRespMessage) Unmarshal(data []byte) error {
	if len(data) < 5 {
		return ErrDataTooShort
	}
	m.RespSeqNum = binary.BigEndian.Uint16(data[0:2])
	m.RespMsgID = binary.BigEndian.Uint16(data[2:4])
	m.Result = data[4]
	return nil
}

// ============================== 0x0002 心跳 ==============================

type HeartbeatMessage struct{}

func (m *HeartbeatMessage) MsgID() uint16                  { return MsgIDHeartbeat }
func (m *HeartbeatMessage) Marshal() ([]byte, error)       { return nil, nil }
func (m *HeartbeatMessage) Unmarshal(data []byte) error     { return nil }

// ============================== 0x0100 终端注册 ==============================

type RegisterMessage struct {
	ProvinceID    uint16 // 省域ID（JT/T 808-2019 表3，2字节）
	CityID        uint16 // 市县域ID（2字节）
	Manufacturer  string // 制造商ID（5字节）
	TerminalModel string // 终端型号（20字节）
	TerminalID    string // 终端ID（7字节）
	PlateColor    byte   // 车牌颜色（1字节）
	PlateNumber   string // 车牌号码（GBK编码，变长）
}

func (m *RegisterMessage) MsgID() uint16 { return MsgIDRegister }

func (m *RegisterMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 0, 50)
	provBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(provBytes, m.ProvinceID)
	buf = append(buf, provBytes...)
	cityBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(cityBytes, m.CityID)
	buf = append(buf, cityBytes...)
	manu := []byte(m.Manufacturer)
	for len(manu) < 5 {
		manu = append(manu, 0)
	}
	buf = append(buf, manu[:5]...)
	model := []byte(m.TerminalModel)
	for len(model) < 20 {
		model = append(model, 0)
	}
	buf = append(buf, model[:20]...)
	tid := []byte(m.TerminalID)
	for len(tid) < 7 {
		tid = append(tid, 0)
	}
	buf = append(buf, tid[:7]...)
	buf = append(buf, m.PlateColor)
	if m.PlateColor != 0 && m.PlateNumber != "" {
		enc := simplifiedchinese.GBK.NewEncoder()
		plateBytes, err := enc.Bytes([]byte(m.PlateNumber))
		if err != nil {
			plateBytes = []byte(m.PlateNumber)
		}
		buf = append(buf, plateBytes...)
	}
	return buf, nil
}

func (m *RegisterMessage) Unmarshal(data []byte) error {
	if len(data) < 36 {
		return ErrDataTooShort
	}
	m.ProvinceID = binary.BigEndian.Uint16(data[0:2])
	m.CityID = binary.BigEndian.Uint16(data[2:4])
	m.Manufacturer = trimNull(data[4:9])
	m.TerminalModel = trimNull(data[9:29])
	m.TerminalID = trimNull(data[29:36])
	if len(data) > 36 {
		m.PlateColor = data[36]
		if len(data) > 37 && m.PlateColor != 0 {
			rawPlate := data[37:]
			dec := simplifiedchinese.GBK.NewDecoder()
			if decoded, err := dec.String(string(rawPlate)); err == nil {
				m.PlateNumber = strings.TrimRight(decoded, "\x00")
			} else {
				m.PlateNumber = strings.TrimRight(string(rawPlate), "\x00")
			}
		}
	}
	return nil
}

// ============================== 0x8100 终端注册应答 ==============================

type RegisterRespMessage struct {
	RespSeqNum  uint16
	Result      byte
	AuthCode    string
}

func (m *RegisterRespMessage) MsgID() uint16 { return MsgIDRegisterResp }

func (m *RegisterRespMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 0, 20)
	seq := make([]byte, 2)
	binary.BigEndian.PutUint16(seq, m.RespSeqNum)
	buf = append(buf, seq...)
	buf = append(buf, m.Result)
	if m.Result == 0 && m.AuthCode != "" {
		buf = append(buf, []byte(m.AuthCode)...)
	}
	return buf, nil
}

func (m *RegisterRespMessage) Unmarshal(data []byte) error {
	if len(data) < 3 {
		return ErrDataTooShort
	}
	m.RespSeqNum = binary.BigEndian.Uint16(data[0:2])
	m.Result = data[2]
	if len(data) > 3 {
		m.AuthCode = string(data[3:])
	}
	return nil
}

// ============================== 0x0102 终端鉴权 ==============================

type AuthMessage struct {
	AuthCode        string
	IMEI            string
	SoftwareVersion string
}

func (m *AuthMessage) MsgID() uint16 { return MsgIDAuth }

func (m *AuthMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 0, len(m.AuthCode)+15+len(m.SoftwareVersion))
	buf = append(buf, []byte(m.AuthCode)...)
	if m.IMEI != "" {
		imei := make([]byte, 15)
		copy(imei, m.IMEI)
		buf = append(buf, imei...)
	}
	if m.SoftwareVersion != "" {
		buf = append(buf, []byte(m.SoftwareVersion)...)
	}
	return buf, nil
}

func (m *AuthMessage) Unmarshal(data []byte) error {
	// JT/T 808-2019: 鉴权码为变长，后续 IMEI 15字节 + 软件版本号
	// 解析策略：如果剩余数据 >= 15 字节，则取最后 15 字节作为 IMEI
	// 如果剩余数据 > 15 字节，则 IMEI 之前的数据为鉴权码，之后为软件版本号
	if len(data) <= 15 {
		// 2011/2013版：只有鉴权码
		m.AuthCode = string(data)
		return nil
	}
	// 2019版：鉴权码 + IMEI(15) + 软件版本号
	imeiStart := len(data) - 15
	m.AuthCode = string(data[:imeiStart])
	m.IMEI = string(data[imeiStart : imeiStart+15])
	// 软件版本号在 IMEI 之后，但长度由终端属性上报决定，此处不解析
	return nil
}

// ============================== 0x0200 位置上报 ==============================

type LocationMessage struct {
	AlarmFlag  uint32  // 报警标志（4字节）
	StatusFlag uint32  // 状态标志（4字节）
	Latitude   float64 // 纬度（度，正值=北纬，负值=南纬）
	Longitude  float64 // 经度（度，正值=东经，负值=西经）
	Altitude   uint16  // 海拔高度（米）
	Speed      uint16  // 速度（0.1km/h）
	Direction  uint16  // 方向（0-359，正北为0顺时针）
	Time       string  // 时间（BCD 6字节: YYMMDDHHmmss）
	ExtraData  []byte  // 附加项（位置附加项）
}

func (m *LocationMessage) MsgID() uint16 { return MsgIDLocation }

func (m *LocationMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 0, 28)
	alarm := make([]byte, 4)
	binary.BigEndian.PutUint32(alarm, m.AlarmFlag)
	buf = append(buf, alarm...)

	status := make([]byte, 4)
	binary.BigEndian.PutUint32(status, m.StatusFlag)
	buf = append(buf, status...)

	// 纬度：取绝对值后乘以缩放因子，协议中为uint32
	lat := uint32(math.Abs(m.Latitude) * CoordScaleFactor)
	if lat > 0xFFFFFFFF {
		lat = 0xFFFFFFFF
	}
	latBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(latBytes, lat)
	buf = append(buf, latBytes...)

	// 经度：取绝对值后乘以缩放因子，协议中为uint32
	lon := uint32(math.Abs(m.Longitude) * CoordScaleFactor)
	if lon > 0xFFFFFFFF {
		lon = 0xFFFFFFFF
	}
	lonBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lonBytes, lon)
	buf = append(buf, lonBytes...)

	alt := make([]byte, 2)
	binary.BigEndian.PutUint16(alt, m.Altitude)
	buf = append(buf, alt...)

	sp := make([]byte, 2)
	binary.BigEndian.PutUint16(sp, m.Speed)
	buf = append(buf, sp...)

	dir := make([]byte, 2)
	binary.BigEndian.PutUint16(dir, m.Direction)
	buf = append(buf, dir...)

	timeBCD, err := StringToBCD(m.Time, 6)
	if err != nil {
		return nil, fmt.Errorf("encode location time: %w", err)
	}
	buf = append(buf, timeBCD...)

	if len(m.ExtraData) > 0 {
		buf = append(buf, m.ExtraData...)
	}

	return buf, nil
}

func (m *LocationMessage) Unmarshal(data []byte) error {
	if len(data) < 28 {
		return ErrDataTooShort
	}
	m.AlarmFlag = binary.BigEndian.Uint32(data[0:4])
	m.StatusFlag = binary.BigEndian.Uint32(data[4:8])

	rawLat := binary.BigEndian.Uint32(data[8:12])
	m.Latitude = float64(rawLat) / CoordScaleFactor
	// Bit 2 of StatusFlag: 0=北纬, 1=南纬
	if m.StatusFlag&0x04 != 0 {
		m.Latitude = -m.Latitude
	}

	rawLon := binary.BigEndian.Uint32(data[12:16])
	m.Longitude = float64(rawLon) / CoordScaleFactor
	// Bit 3 of StatusFlag: 0=东经, 1=西经
	if m.StatusFlag&0x08 != 0 {
		m.Longitude = -m.Longitude
	}

	m.Altitude = binary.BigEndian.Uint16(data[16:18])
	m.Speed = binary.BigEndian.Uint16(data[18:20])
	m.Direction = binary.BigEndian.Uint16(data[20:22])
	m.Time = BCDToStringSafe(data[22:28])

	if len(data) > 28 {
		m.ExtraData = make([]byte, len(data)-28)
		copy(m.ExtraData, data[28:])
	}

	return nil
}

// ============================== 0x0704 批量位置上报 ==============================

type LocationBatchMessage struct {
	Type    byte
	Items   []*LocationMessage
}

func (m *LocationBatchMessage) MsgID() uint16 { return MsgIDLocationBatch }

func (m *LocationBatchMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 0, 256)
	buf = append(buf, m.Type)
	count := make([]byte, 2)
	binary.BigEndian.PutUint16(count, uint16(len(m.Items)))
	buf = append(buf, count...)

	for _, loc := range m.Items {
		locData, err := loc.Marshal()
		if err != nil {
			return nil, err
		}
		// JT/T 808-2019: 位置项长度为2字节
		lenBytes := make([]byte, 2)
		binary.BigEndian.PutUint16(lenBytes, uint16(len(locData)))
		buf = append(buf, lenBytes...)
		buf = append(buf, locData...)
	}

	return buf, nil
}

func (m *LocationBatchMessage) Unmarshal(data []byte) error {
	if len(data) < 3 {
		return ErrDataTooShort
	}
	m.Type = data[0]
	count := binary.BigEndian.Uint16(data[1:3])
	offset := 3
	for i := 0; i < int(count); i++ {
		if offset+2 > len(data) {
			return fmt.Errorf("batch location: item %d length out of range", i)
		}
		// JT/T 808-2019: 位置项长度为2字节
		itemLen := int(binary.BigEndian.Uint16(data[offset : offset+2]))
		offset += 2
		if offset+itemLen > len(data) {
			return fmt.Errorf("batch location: item %d data too short", i)
		}
		loc := &LocationMessage{}
		if err := loc.Unmarshal(data[offset : offset+itemLen]); err != nil {
			return fmt.Errorf("batch location item %d: %w", i, err)
		}
		m.Items = append(m.Items, loc)
		offset += itemLen
	}
	return nil
}

// ============================== 0x8103 设置参数 ==============================

type CommandMessage struct {
	Params []ParamItem
}

type ParamItem struct {
	ParamID uint32
	ParamLen byte
	ParamValue []byte
}

func (m *CommandMessage) MsgID() uint16 { return MsgIDCommand }

func (m *CommandMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 0, 256)
	for _, p := range m.Params {
		idBytes := make([]byte, 4)
		binary.BigEndian.PutUint32(idBytes, p.ParamID)
		buf = append(buf, idBytes...)
		buf = append(buf, p.ParamLen)
		buf = append(buf, p.ParamValue...)
	}
	return buf, nil
}

func (m *CommandMessage) Unmarshal(data []byte) error {
	offset := 0
	for offset+5 <= len(data) {
		var p ParamItem
		p.ParamID = binary.BigEndian.Uint32(data[offset : offset+4])
		p.ParamLen = data[offset+4]
		offset += 5
		if offset+int(p.ParamLen) > len(data) {
			return fmt.Errorf("param 0x%04X: data too short", p.ParamID)
		}
		p.ParamValue = make([]byte, p.ParamLen)
		copy(p.ParamValue, data[offset:offset+int(p.ParamLen)])
		offset += int(p.ParamLen)
		m.Params = append(m.Params, p)
	}
	return nil
}

// ============================== 0x0103 设置参数应答 ==============================

type CommandRespMessage struct {
	RespSeqNum uint16
	Result     byte
	ParamCount byte
	Params     []uint32
}

func (m *CommandRespMessage) MsgID() uint16 { return MsgIDCommandResp }

func (m *CommandRespMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 0, 20)
	seq := make([]byte, 2)
	binary.BigEndian.PutUint16(seq, m.RespSeqNum)
	buf = append(buf, seq...)
	buf = append(buf, m.Result)
	buf = append(buf, m.ParamCount)
	for _, id := range m.Params {
		idBytes := make([]byte, 4)
		binary.BigEndian.PutUint32(idBytes, id)
		buf = append(buf, idBytes...)
	}
	return buf, nil
}

func (m *CommandRespMessage) Unmarshal(data []byte) error {
	if len(data) < 4 {
		return ErrDataTooShort
	}
	m.RespSeqNum = binary.BigEndian.Uint16(data[0:2])
	m.Result = data[2]
	m.ParamCount = data[3]
	for i := 0; i < int(m.ParamCount); i++ {
		offset := 4 + i*4
		if offset+4 > len(data) {
			return ErrDataTooShort
		}
		m.Params = append(m.Params, binary.BigEndian.Uint32(data[offset:offset+4]))
	}
	return nil
}

// ============================== 0x8104 查询参数 ==============================

type ParamQueryMessage struct {
	ParamIDs []uint32
}

func (m *ParamQueryMessage) MsgID() uint16 { return MsgIDParamQuery }

func (m *ParamQueryMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 0, 20)
	for _, id := range m.ParamIDs {
		idBytes := make([]byte, 4)
		binary.BigEndian.PutUint32(idBytes, id)
		buf = append(buf, idBytes...)
	}
	return buf, nil
}

func (m *ParamQueryMessage) Unmarshal(data []byte) error {
	for i := 0; i+4 <= len(data); i += 4 {
		m.ParamIDs = append(m.ParamIDs, binary.BigEndian.Uint32(data[i:i+4]))
	}
	return nil
}

// ============================== 0x0104 查询参数应答 ==============================

type ParamRespMessage struct {
	Params []ParamItem
}

func (m *ParamRespMessage) MsgID() uint16 { return MsgIDParamResp }

func (m *ParamRespMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 0, 256)
	for _, p := range m.Params {
		idBytes := make([]byte, 4)
		binary.BigEndian.PutUint32(idBytes, p.ParamID)
		buf = append(buf, idBytes...)
		buf = append(buf, p.ParamLen)
		buf = append(buf, p.ParamValue...)
	}
	return buf, nil
}

func (m *ParamRespMessage) Unmarshal(data []byte) error {
	offset := 0
	for offset+5 <= len(data) {
		var p ParamItem
		p.ParamID = binary.BigEndian.Uint32(data[offset : offset+4])
		p.ParamLen = data[offset+4]
		offset += 5
		if offset+int(p.ParamLen) > len(data) {
			return fmt.Errorf("param resp 0x%04X: data too short", p.ParamID)
		}
		p.ParamValue = make([]byte, p.ParamLen)
		copy(p.ParamValue, data[offset:offset+int(p.ParamLen)])
		offset += int(p.ParamLen)
		m.Params = append(m.Params, p)
	}
	return nil
}

// ============================== 0x8105 终端控制 ==============================

type TerminalCtrlMessage struct {
	Command byte
	Param   string
}

func (m *TerminalCtrlMessage) MsgID() uint16 { return MsgIDTerminalCtrl }

func (m *TerminalCtrlMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 0, 20)
	buf = append(buf, m.Command)
	if m.Param != "" {
		buf = append(buf, []byte(m.Param)...)
	}
	return buf, nil
}

func (m *TerminalCtrlMessage) Unmarshal(data []byte) error {
	if len(data) < 1 {
		return ErrDataTooShort
	}
	m.Command = data[0]
	if len(data) > 1 {
		m.Param = string(data[1:])
	}
	return nil
}

// ============================== 0x8201 查询位置 ==============================

type LocationQueryMessage struct{}

func (m *LocationQueryMessage) MsgID() uint16                  { return MsgIDLocationQuery }
func (m *LocationQueryMessage) Marshal() ([]byte, error)       { return nil, nil }
func (m *LocationQueryMessage) Unmarshal(data []byte) error     { return nil }

// ============================== 0x8202 临时位置跟踪 ==============================

type TempLocationTrackMessage struct {
	Interval uint16
	Validity uint16
}

func (m *TempLocationTrackMessage) MsgID() uint16 { return MsgIDTempLocationTrack }

func (m *TempLocationTrackMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint16(buf[0:2], m.Interval)
	binary.BigEndian.PutUint16(buf[2:4], m.Validity)
	return buf, nil
}

func (m *TempLocationTrackMessage) Unmarshal(data []byte) error {
	if len(data) < 4 {
		return ErrDataTooShort
	}
	m.Interval = binary.BigEndian.Uint16(data[0:2])
	m.Validity = binary.BigEndian.Uint16(data[2:4])
	return nil
}

// ============================== 0x8203 人工确认报警 ==============================

type ManualAlarmConfirmMessage struct {
	SeqNum  uint16
	AlarmID uint16
}

func (m *ManualAlarmConfirmMessage) MsgID() uint16 { return MsgIDManualAlarmConfirm }

func (m *ManualAlarmConfirmMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint16(buf[0:2], m.SeqNum)
	binary.BigEndian.PutUint16(buf[2:4], m.AlarmID)
	return buf, nil
}

func (m *ManualAlarmConfirmMessage) Unmarshal(data []byte) error {
	if len(data) < 4 {
		return ErrDataTooShort
	}
	m.SeqNum = binary.BigEndian.Uint16(data[0:2])
	m.AlarmID = binary.BigEndian.Uint16(data[2:4])
	return nil
}

// ============================== 0x8300 文本下发 ==============================

type TextSendMessage struct {
	Flag     byte
	Content  string
}

func (m *TextSendMessage) MsgID() uint16 { return MsgIDTextSend }

func (m *TextSendMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 0, 256)
	buf = append(buf, m.Flag)
	enc := simplifiedchinese.GBK.NewEncoder()
	content, err := enc.Bytes([]byte(m.Content))
	if err != nil {
		content = []byte(m.Content)
	}
	buf = append(buf, content...)
	return buf, nil
}

func (m *TextSendMessage) Unmarshal(data []byte) error {
	if len(data) < 1 {
		return ErrDataTooShort
	}
	m.Flag = data[0]
	if len(data) > 1 {
		dec := simplifiedchinese.GBK.NewDecoder()
		if decoded, err := dec.String(string(data[1:])); err == nil {
			m.Content = decoded
		} else {
			m.Content = string(data[1:])
		}
	}
	return nil
}

// ============================== 0x8400 超速报警设置 ==============================

type OverspeedSetMessage struct {
	SeqNum    uint16
	Speed     uint16
	Duration  uint16
}

func (m *OverspeedSetMessage) MsgID() uint16 { return MsgIDOverspeedSet }

func (m *OverspeedSetMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 6)
	binary.BigEndian.PutUint16(buf[0:2], m.SeqNum)
	binary.BigEndian.PutUint16(buf[2:4], m.Speed)
	binary.BigEndian.PutUint16(buf[4:6], m.Duration)
	return buf, nil
}

func (m *OverspeedSetMessage) Unmarshal(data []byte) error {
	if len(data) < 6 {
		return ErrDataTooShort
	}
	m.SeqNum = binary.BigEndian.Uint16(data[0:2])
	m.Speed = binary.BigEndian.Uint16(data[2:4])
	m.Duration = binary.BigEndian.Uint16(data[4:6])
	return nil
}

// ============================== 0x8401 疲劳驾驶设置 ==============================

type FatigueDriveSetMessage struct {
	SeqNum       uint16
	DayMaxDrive  uint16
	DayMinRest   uint16
	MaxDrive     uint16
	MinRest      uint16
}

func (m *FatigueDriveSetMessage) MsgID() uint16 { return MsgIDFatigueDriveSet }

func (m *FatigueDriveSetMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 10)
	binary.BigEndian.PutUint16(buf[0:2], m.SeqNum)
	binary.BigEndian.PutUint16(buf[2:4], m.DayMaxDrive)
	binary.BigEndian.PutUint16(buf[4:6], m.DayMinRest)
	binary.BigEndian.PutUint16(buf[6:8], m.MaxDrive)
	binary.BigEndian.PutUint16(buf[8:10], m.MinRest)
	return buf, nil
}

func (m *FatigueDriveSetMessage) Unmarshal(data []byte) error {
	if len(data) < 10 {
		return ErrDataTooShort
	}
	m.SeqNum = binary.BigEndian.Uint16(data[0:2])
	m.DayMaxDrive = binary.BigEndian.Uint16(data[2:4])
	m.DayMinRest = binary.BigEndian.Uint16(data[4:6])
	m.MaxDrive = binary.BigEndian.Uint16(data[6:8])
	m.MinRest = binary.BigEndian.Uint16(data[8:10])
	return nil
}

// ============================== 0x8500 车辆控制 ==============================

type VehicleControlMessage struct {
	ControlFlag byte
}

func (m *VehicleControlMessage) MsgID() uint16 { return MsgIDVehicleControl }

func (m *VehicleControlMessage) Marshal() ([]byte, error) {
	return []byte{m.ControlFlag}, nil
}

func (m *VehicleControlMessage) Unmarshal(data []byte) error {
	if len(data) < 1 {
		return ErrDataTooShort
	}
	m.ControlFlag = data[0]
	return nil
}

// ============================== 0x8600 圆形区域设置 ==============================

type CircularAreaSetMessage struct {
	AreaID   uint32
	Attr     uint16
	CenterLat float64
	CenterLon float64
	Radius   uint32
	StartTime string
	EndTime   string
	SpeedLimit uint16
}

func (m *CircularAreaSetMessage) MsgID() uint16 { return MsgIDCircularAreaSet }

func (m *CircularAreaSetMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 0, 40)
	idBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(idBytes, m.AreaID)
	buf = append(buf, idBytes...)

	attrBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(attrBytes, m.Attr)
	buf = append(buf, attrBytes...)

	latBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(latBytes, uint32(m.CenterLat*CoordScaleFactor))
	buf = append(buf, latBytes...)

	lonBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lonBytes, uint32(m.CenterLon*CoordScaleFactor))
	buf = append(buf, lonBytes...)

	rBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(rBytes, m.Radius)
	buf = append(buf, rBytes...)

	startBCD, _ := StringToBCD(m.StartTime, 3)
	buf = append(buf, startBCD...)
	endBCD, _ := StringToBCD(m.EndTime, 3)
	buf = append(buf, endBCD...)

	if m.Attr&0x02 != 0 {
		slBytes := make([]byte, 2)
		binary.BigEndian.PutUint16(slBytes, m.SpeedLimit)
		buf = append(buf, slBytes...)
	}

	return buf, nil
}

func (m *CircularAreaSetMessage) Unmarshal(data []byte) error {
	// 最小长度：4(id)+2(attr)+4(lat)+4(lon)+4(radius)+3(start)+3(end) = 24
	if len(data) < 24 {
		return ErrDataTooShort
	}
	m.AreaID = binary.BigEndian.Uint32(data[0:4])
	m.Attr = binary.BigEndian.Uint16(data[4:6])
	m.CenterLat = float64(binary.BigEndian.Uint32(data[6:10])) / CoordScaleFactor
	m.CenterLon = float64(binary.BigEndian.Uint32(data[10:14])) / CoordScaleFactor
	m.Radius = binary.BigEndian.Uint32(data[14:18])
	m.StartTime = BCDToStringSafe(data[18:21])
	m.EndTime = BCDToStringSafe(data[21:24])
	if len(data) >= 26 && m.Attr&0x02 != 0 {
		m.SpeedLimit = binary.BigEndian.Uint16(data[24:26])
	}
	return nil
}

// ============================== 0x8602 矩形区域设置 ==============================

type RectAreaSetMessage struct {
	AreaID    uint32
	Attr      uint16
	UpperLat  float64
	UpperLon  float64
	LowerLat  float64
	LowerLon  float64
	StartTime string
	EndTime   string
	SpeedLimit uint16
}

func (m *RectAreaSetMessage) MsgID() uint16 { return MsgIDRectAreaSet }

func (m *RectAreaSetMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 0, 40)
	idBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(idBytes, m.AreaID)
	buf = append(buf, idBytes...)

	attrBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(attrBytes, m.Attr)
	buf = append(buf, attrBytes...)

	ulLat := make([]byte, 4)
	binary.BigEndian.PutUint32(ulLat, uint32(m.UpperLat*CoordScaleFactor))
	buf = append(buf, ulLat...)

	ulLon := make([]byte, 4)
	binary.BigEndian.PutUint32(ulLon, uint32(m.UpperLon*CoordScaleFactor))
	buf = append(buf, ulLon...)

	llLat := make([]byte, 4)
	binary.BigEndian.PutUint32(llLat, uint32(m.LowerLat*CoordScaleFactor))
	buf = append(buf, llLat...)

	llLon := make([]byte, 4)
	binary.BigEndian.PutUint32(llLon, uint32(m.LowerLon*CoordScaleFactor))
	buf = append(buf, llLon...)

	startBCD, _ := StringToBCD(m.StartTime, 3)
	buf = append(buf, startBCD...)
	endBCD, _ := StringToBCD(m.EndTime, 3)
	buf = append(buf, endBCD...)

	if m.Attr&0x02 != 0 {
		slBytes := make([]byte, 2)
		binary.BigEndian.PutUint16(slBytes, m.SpeedLimit)
		buf = append(buf, slBytes...)
	}

	return buf, nil
}

func (m *RectAreaSetMessage) Unmarshal(data []byte) error {
	// 最小长度：4(id)+2(attr)+4*4(4个坐标)+3(start)+3(end) = 28
	if len(data) < 28 {
		return ErrDataTooShort
	}
	m.AreaID = binary.BigEndian.Uint32(data[0:4])
	m.Attr = binary.BigEndian.Uint16(data[4:6])
	m.UpperLat = float64(binary.BigEndian.Uint32(data[6:10])) / CoordScaleFactor
	m.UpperLon = float64(binary.BigEndian.Uint32(data[10:14])) / CoordScaleFactor
	m.LowerLat = float64(binary.BigEndian.Uint32(data[14:18])) / CoordScaleFactor
	m.LowerLon = float64(binary.BigEndian.Uint32(data[18:22])) / CoordScaleFactor
	m.StartTime = BCDToStringSafe(data[22:25])
	m.EndTime = BCDToStringSafe(data[25:28])
	if len(data) >= 30 && m.Attr&0x02 != 0 {
		m.SpeedLimit = binary.BigEndian.Uint16(data[28:30])
	}
	return nil
}

// ============================== 0x0702 驾驶员身份 ==============================

type DriverIDMessage struct {
	DriverName   string
	DriverID     string
	Licence      string
	CertifyOrg   string
}

func (m *DriverIDMessage) MsgID() uint16 { return MsgIDDriverID }

func (m *DriverIDMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 0, 80)
	enc := simplifiedchinese.GBK.NewEncoder()
	nameBytes, _ := enc.Bytes([]byte(m.DriverName))
	buf = append(buf, byte(len(nameBytes)))
	buf = append(buf, nameBytes...)

	idBytes := []byte(m.DriverID)
	for len(idBytes) < 20 {
		idBytes = append(idBytes, 0)
	}
	buf = append(buf, idBytes[:20]...)

	licBytes := []byte(m.Licence)
	for len(licBytes) < 40 {
		licBytes = append(licBytes, 0)
	}
	buf = append(buf, licBytes[:40]...)

	orgBytes, _ := enc.Bytes([]byte(m.CertifyOrg))
	for len(orgBytes) < 20 {
		orgBytes = append(orgBytes, 0)
	}
	buf = append(buf, orgBytes[:20]...)

	return buf, nil
}

func (m *DriverIDMessage) Unmarshal(data []byte) error {
	if len(data) < 1 {
		return ErrDataTooShort
	}
	nameLen := int(data[0])
	offset := 1
	if offset+nameLen > len(data) {
		return ErrDataTooShort
	}
	dec := simplifiedchinese.GBK.NewDecoder()
	if decoded, err := dec.String(string(data[offset : offset+nameLen])); err == nil {
		m.DriverName = decoded
	} else {
		m.DriverName = string(data[offset : offset+nameLen])
	}
	offset += nameLen

	if offset+20 > len(data) {
		return ErrDataTooShort
	}
	m.DriverID = trimNull(data[offset : offset+20])
	offset += 20

	if offset+40 > len(data) {
		return ErrDataTooShort
	}
	m.Licence = trimNull(data[offset : offset+40])
	offset += 40

	if offset+20 > len(data) {
		return ErrDataTooShort
	}
	if decoded, err := dec.String(string(data[offset : offset+20])); err == nil {
		m.CertifyOrg = trimNull([]byte(decoded))
	} else {
		m.CertifyOrg = trimNull(data[offset : offset+20])
	}
	return nil
}

// ============================== 0x0705 CAN数据 ==============================

type CanDataMessage struct {
	CanItems []CanItem
}

type CanItem struct {
	CanTime  uint32
	CanID    uint32
	CanData  []byte
}

func (m *CanDataMessage) MsgID() uint16 { return MsgIDCanData }

func (m *CanDataMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 0, 256)
	count := byte(len(m.CanItems))
	buf = append(buf, count)
	for _, item := range m.CanItems {
		timeBytes := make([]byte, 4)
		binary.BigEndian.PutUint32(timeBytes, item.CanTime)
		buf = append(buf, timeBytes...)

		canIDBytes := make([]byte, 4)
		binary.BigEndian.PutUint32(canIDBytes, item.CanID)
		buf = append(buf, canIDBytes...)

		buf = append(buf, byte(len(item.CanData)))
		buf = append(buf, item.CanData...)
	}
	return buf, nil
}

func (m *CanDataMessage) Unmarshal(data []byte) error {
	if len(data) < 1 {
		return ErrDataTooShort
	}
	count := int(data[0])
	offset := 1
	for i := 0; i < count; i++ {
		if offset+9 > len(data) {
			return ErrDataTooShort
		}
		var item CanItem
		item.CanTime = binary.BigEndian.Uint32(data[offset : offset+4])
		item.CanID = binary.BigEndian.Uint32(data[offset+4 : offset+8])
		dataLen := int(data[offset+8])
		offset += 9
		if offset+dataLen > len(data) {
			return ErrDataTooShort
		}
		item.CanData = make([]byte, dataLen)
		copy(item.CanData, data[offset:offset+dataLen])
		offset += dataLen
		m.CanItems = append(m.CanItems, item)
	}
	return nil
}

// ============================== 0x8107 终端属性查询 ==============================

type TerminalPropQueryMessage struct{}

func (m *TerminalPropQueryMessage) MsgID() uint16                  { return MsgIDTerminalPropQuery }
func (m *TerminalPropQueryMessage) Marshal() ([]byte, error)       { return nil, nil }
func (m *TerminalPropQueryMessage) Unmarshal(data []byte) error     { return nil }

// ============================== 0x8108 终端升级 ==============================

type TerminalUpgradeMessage struct {
	UpgradeType byte
	Manufacturer string
	Model        string
	Version      string
	URL          string
}

func (m *TerminalUpgradeMessage) MsgID() uint16 { return MsgIDTerminalUpgrade }

func (m *TerminalUpgradeMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 0, 100)
	buf = append(buf, m.UpgradeType)
	manu := []byte(m.Manufacturer)
	for len(manu) < 5 {
		manu = append(manu, 0)
	}
	buf = append(buf, manu[:5]...)
	model := []byte(m.Model)
	for len(model) < 20 {
		model = append(model, 0)
	}
	buf = append(buf, model[:20]...)
	buf = append(buf, byte(len(m.Version)))
	buf = append(buf, []byte(m.Version)...)
	buf = append(buf, []byte(m.URL)...)
	return buf, nil
}

func (m *TerminalUpgradeMessage) Unmarshal(data []byte) error {
	if len(data) < 27 {
		return ErrDataTooShort
	}
	m.UpgradeType = data[0]
	m.Manufacturer = trimNull(data[1:6])
	m.Model = trimNull(data[6:26])
	verLen := int(data[26])
	offset := 27
	if offset+verLen > len(data) {
		return ErrDataTooShort
	}
	m.Version = string(data[offset : offset+verLen])
	offset += verLen
	if offset < len(data) {
		m.URL = string(data[offset:])
	}
	return nil
}

// ============================== 0x0801 多媒体事件 ==============================

// MultimediaMessage 0x0801 多媒体事件信息上报
// JT/T 808-2019 表46: 多媒体ID(4) + 多媒体类型(1) + 编码格式(1) + 事件项编码(1) + 通道号(1) = 8字节
type MultimediaMessage struct {
	MultimediaID   uint32 // 多媒体ID
	MultimediaType byte   // 多媒体类型：0=图像 1=音频 2=视频
	Format         byte   // 编码格式
	EventCode      byte   // 事件项编码
	ChannelID      byte   // 通道号
}

func (m *MultimediaMessage) MsgID() uint16 { return MsgIDMultimedia }

func (m *MultimediaMessage) Marshal() ([]byte, error) {
	// JT/T 808-2019: 8字节
	buf := make([]byte, 8)
	binary.BigEndian.PutUint32(buf[0:4], m.MultimediaID)
	buf[4] = m.MultimediaType
	buf[5] = m.Format
	buf[6] = m.EventCode
	buf[7] = m.ChannelID
	return buf, nil
}

func (m *MultimediaMessage) Unmarshal(data []byte) error {
	if len(data) < 8 {
		return ErrDataTooShort
	}
	m.MultimediaID = binary.BigEndian.Uint32(data[0:4])
	m.MultimediaType = data[4]
	m.Format = data[5]
	m.EventCode = data[6]
	m.ChannelID = data[7]
	return nil
}

// ============================== 0x8801 摄像头立即拍摄 ==============================

type PhotoCommandMessage struct {
	SeqNum     uint16
	ChannelID  byte
	Command    byte
	Interval   uint16
	Count      byte
	Resolution byte
	Quality    byte
	Luminance  byte
	Contrast   byte
	Saturation byte
	Chroma     byte
}

func (m *PhotoCommandMessage) MsgID() uint16 { return MsgIDPhotoCommand }

func (m *PhotoCommandMessage) Marshal() ([]byte, error) {
	// JT/T 808-2019 表56: 12字节(2011版) 或 13字节(2019版含色度)
	buf := make([]byte, 13)
	binary.BigEndian.PutUint16(buf[0:2], m.SeqNum)
	buf[2] = m.ChannelID
	buf[3] = m.Command
	binary.BigEndian.PutUint16(buf[4:6], m.Interval)
	buf[6] = m.Count
	buf[7] = m.Resolution
	buf[8] = m.Quality
	buf[9] = m.Luminance
	buf[10] = m.Contrast
	buf[11] = m.Saturation
	buf[12] = m.Chroma
	return buf, nil
}

func (m *PhotoCommandMessage) Unmarshal(data []byte) error {
	if len(data) < 12 {
		return ErrDataTooShort
	}
	m.SeqNum = binary.BigEndian.Uint16(data[0:2])
	m.ChannelID = data[2]
	m.Command = data[3]
	m.Interval = binary.BigEndian.Uint16(data[4:6])
	m.Count = data[6]
	m.Resolution = data[7]
	m.Quality = data[8]
	m.Luminance = data[9]
	m.Contrast = data[10]
	m.Saturation = data[11]
	if len(data) > 12 {
		m.Chroma = data[12]
	}
	return nil
}

// ============================== 0x8301 事件设置 ==============================

type EventSetMessage struct {
	Count byte
	Items []EventItem
}

type EventItem struct {
	EventID byte
	Content string
}

func (m *EventSetMessage) MsgID() uint16 { return MsgIDEventSet }

func (m *EventSetMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 0, 100)
	buf = append(buf, byte(len(m.Items)))
	for _, item := range m.Items {
		buf = append(buf, item.EventID)
		enc := simplifiedchinese.GBK.NewEncoder()
		contentBytes, _ := enc.Bytes([]byte(item.Content))
		buf = append(buf, byte(len(contentBytes)))
		buf = append(buf, contentBytes...)
	}
	return buf, nil
}

func (m *EventSetMessage) Unmarshal(data []byte) error {
	if len(data) < 1 {
		return ErrDataTooShort
	}
	count := int(data[0])
	offset := 1
	dec := simplifiedchinese.GBK.NewDecoder()
	for i := 0; i < count; i++ {
		if offset+2 > len(data) {
			return ErrDataTooShort
		}
		var item EventItem
		item.EventID = data[offset]
		contentLen := int(data[offset+1])
		offset += 2
		if offset+contentLen > len(data) {
			return ErrDataTooShort
		}
		if decoded, err := dec.String(string(data[offset : offset+contentLen])); err == nil {
			item.Content = decoded
		} else {
			item.Content = string(data[offset : offset+contentLen])
		}
		offset += contentLen
		m.Items = append(m.Items, item)
	}
	return nil
}

// ============================== 0x0301 事件应答 ==============================

type EventRespMessage struct {
	SeqNum  uint16
	EventID byte
}

func (m *EventRespMessage) MsgID() uint16 { return MsgIDEventResp }

func (m *EventRespMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 3)
	binary.BigEndian.PutUint16(buf[0:2], m.SeqNum)
	buf[2] = m.EventID
	return buf, nil
}

func (m *EventRespMessage) Unmarshal(data []byte) error {
	if len(data) < 3 {
		return ErrDataTooShort
	}
	m.SeqNum = binary.BigEndian.Uint16(data[0:2])
	m.EventID = data[2]
	return nil
}

// ============================== 0x0900 报警附件 ==============================
// JT/T 808-2019: 0x0900 为报警附件信息上报，不是简单报警
// 报警主要通过 0x0200 位置上报的 AlarmFlag 字段传递

type AlarmMessage struct {
	VehicleColor byte   // 车牌颜色
	VehiclePlate string // 车牌号码（21字节）
	AlarmFlag    uint16 // 报警标志（2字节，0x0900特有）
	WaterLevel   byte   // 水位（1字节）
	ExtraData    []byte // 附加数据
}

func (m *AlarmMessage) MsgID() uint16 { return MsgIDAlarm }

func (m *AlarmMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 0, 30)
	buf = append(buf, m.VehicleColor)
	plate := []byte(m.VehiclePlate)
	for len(plate) < 21 {
		plate = append(plate, 0)
	}
	buf = append(buf, plate[:21]...)

	flagBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(flagBytes, m.AlarmFlag)
	buf = append(buf, flagBytes...)
	buf = append(buf, m.WaterLevel)
	if len(m.ExtraData) > 0 {
		buf = append(buf, m.ExtraData...)
	}
	return buf, nil
}

func (m *AlarmMessage) Unmarshal(data []byte) error {
	if len(data) < 25 {
		return ErrDataTooShort
	}
	m.VehicleColor = data[0]
	plate := data[1:22]
	for len(plate) > 0 && plate[len(plate)-1] == 0 {
		plate = plate[:len(plate)-1]
	}
	m.VehiclePlate = string(plate)
	m.AlarmFlag = binary.BigEndian.Uint16(data[22:24])
	m.WaterLevel = data[24]
	if len(data) > 25 {
		m.ExtraData = make([]byte, len(data)-25)
		copy(m.ExtraData, data[25:])
	}
	return nil
}

// ============================== 0x0400 超速报警 ==============================

type OverspeedAlarmMessage struct {
	AlarmFlag uint16
	ExtraData []byte
}

func (m *OverspeedAlarmMessage) MsgID() uint16 { return MsgIDOverspeedAlarm }

func (m *OverspeedAlarmMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 0, 10)
	flagBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(flagBytes, m.AlarmFlag)
	buf = append(buf, flagBytes...)
	if len(m.ExtraData) > 0 {
		buf = append(buf, m.ExtraData...)
	}
	return buf, nil
}

func (m *OverspeedAlarmMessage) Unmarshal(data []byte) error {
	if len(data) < 2 {
		return ErrDataTooShort
	}
	m.AlarmFlag = binary.BigEndian.Uint16(data[0:2])
	if len(data) > 2 {
		m.ExtraData = make([]byte, len(data)-2)
		copy(m.ExtraData, data[2:])
	}
	return nil
}

// ============================== 0x0401 疲劳驾驶报警 ==============================

type FatigueDriveAlarmMessage struct {
	AlarmFlag uint16
	ExtraData []byte
}

func (m *FatigueDriveAlarmMessage) MsgID() uint16 { return MsgIDFatigueDriveAlarm }

func (m *FatigueDriveAlarmMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 0, 10)
	flagBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(flagBytes, m.AlarmFlag)
	buf = append(buf, flagBytes...)
	if len(m.ExtraData) > 0 {
		buf = append(buf, m.ExtraData...)
	}
	return buf, nil
}

func (m *FatigueDriveAlarmMessage) Unmarshal(data []byte) error {
	if len(data) < 2 {
		return ErrDataTooShort
	}
	m.AlarmFlag = binary.BigEndian.Uint16(data[0:2])
	if len(data) > 2 {
		m.ExtraData = make([]byte, len(data)-2)
		copy(m.ExtraData, data[2:])
	}
	return nil
}

// ============================== 辅助函数 ==============================

// unused import guard
var _ = math.Abs
var _ = strings.TrimRight