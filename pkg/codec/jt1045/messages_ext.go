package jt1045

import (
	"encoding/binary"
	"fmt"
	"math"
)

// ==================== JT/T 1045-2018 完整消息ID ====================

const (
	MsgIDADASAlarm      uint16 = 0x0901 // ADAS报警
	MsgIDADASAlarmResp  uint16 = 0x8901 // ADAS报警应答
	MsgIDADASData       uint16 = 0x0902 // ADAS数据
	MsgIDADASDataResp   uint16 = 0x8902 // ADAS数据应答
	MsgIDDSMAlarm       uint16 = 0x1205 // DSM报警
	MsgIDDSMAlarmResp   uint16 = 0x9205 // DSM报警应答
	MsgIDDSMData        uint16 = 0x1207 // DSM数据
	MsgIDDSMDataResp    uint16 = 0x9207 // DSM数据应答
	MsgIDBSDAlarm       uint16 = 0x1209 // 盲区报警(BSD)
	MsgIDBSDAlarmResp   uint16 = 0x9209 // 盲区报警应答
	MsgIDBSDData        uint16 = 0x120A // 盲区数据
	MsgIDTireAlarm      uint16 = 0x120B // 胎压报警
	MsgIDTireAlarmResp  uint16 = 0x920B // 胎压报警应答
	MsgIDTireData       uint16 = 0x120C // 胎压数据
	MsgIDADASParam      uint16 = 0x8903 // ADAS参数设置
	MsgIDADASParamResp  uint16 = 0x0903 // ADAS参数应答
	MsgIDDSMParam       uint16 = 0x9206 // DSM参数设置
	MsgIDDSMParamResp   uint16 = 0x1206 // DSM参数应答
	MsgIDBSDParam       uint16 = 0x920A // BSD参数设置
	MsgIDBSDParamResp   uint16 = 0x120D // BSD参数应答
	MsgIDTireParam      uint16 = 0x920C // 胎压参数设置
	MsgIDTireParamResp  uint16 = 0x120E // 胎压参数应答
	MsgIDADASStatus     uint16 = 0x0904 // ADAS状态上报
	MsgIDDSMStatus      uint16 = 0x1208 // DSM状态上报
	MsgIDBSDStatus      uint16 = 0x120F // BSD状态上报
	MsgIDTireStatus     uint16 = 0x1210 // 胎压状态上报
)

// MsgInfo1045 1045消息信息
type MsgInfo1045 struct {
	ID   uint16
	Name string
	Dir  string
}

// AllMessages1045 1045协议全部消息列表
var AllMessages1045 = []MsgInfo1045{
	{0x0901, "ADAS报警", "up"},
	{0x8901, "ADAS报警应答", "down"},
	{0x0902, "ADAS数据", "up"},
	{0x8902, "ADAS数据应答", "down"},
	{0x8903, "ADAS参数设置", "down"},
	{0x0903, "ADAS参数应答", "up"},
	{0x0904, "ADAS状态上报", "up"},
	{0x1205, "DSM报警", "up"},
	{0x9205, "DSM报警应答", "down"},
	{0x9206, "DSM参数设置", "down"},
	{0x1206, "DSM参数应答", "up"},
	{0x1207, "DSM数据", "up"},
	{0x9207, "DSM数据应答", "down"},
	{0x1208, "DSM状态上报", "up"},
	{0x1209, "盲区报警(BSD)", "up"},
	{0x9209, "盲区报警应答", "down"},
	{0x920A, "BSD参数设置", "down"},
	{0x120A, "盲区数据", "up"},
	{0x120D, "BSD参数应答", "up"},
	{0x120F, "BSD状态上报", "up"},
	{0x120B, "胎压报警", "up"},
	{0x920B, "胎压报警应答", "down"},
	{0x920C, "胎压参数设置", "down"},
	{0x120C, "胎压数据", "up"},
	{0x120E, "胎压参数应答", "up"},
	{0x1210, "胎压状态上报", "up"},
}

// ADASDataMessage ADAS数据
type ADASDataMessage struct {
	VehicleColor byte
	VehiclePlate string
	AlarmTime    string
	AlarmType    uint16
	AlarmLevel   byte
	Lat          float64
	Lon          float64
	Speed        uint16
	Direction    uint16
	VehicleNo    string // 前车车牌号
	VehicleSpeed uint16 // 前车速度
	Distance     uint16 // 车距
	DataItems    []byte // 附加数据项
}

func (m *ADASDataMessage) MsgID() uint16 { return MsgIDADASData }

func (m *ADASDataMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 0, 60)
	buf = append(buf, m.VehicleColor)
	plate := []byte(m.VehiclePlate)
	for len(plate) < 21 {
		plate = append(plate, 0)
	}
	buf = append(buf, plate[:21]...)
	for len(m.AlarmTime) < 12 {
		m.AlarmTime = "0" + m.AlarmTime
	}
	for i := 0; i < 6; i++ {
		buf = append(buf, ((m.AlarmTime[i*2]-'0')<<4)|(m.AlarmTime[i*2+1]-'0'))
	}
	typeBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(typeBytes, m.AlarmType)
	buf = append(buf, typeBytes...)
	buf = append(buf, m.AlarmLevel)
	latBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(latBytes, uint32(math.Abs(m.Lat)*1000000))
	buf = append(buf, latBytes...)
	lonBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lonBytes, uint32(math.Abs(m.Lon)*1000000))
	buf = append(buf, lonBytes...)
	speedBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(speedBytes, m.Speed)
	buf = append(buf, speedBytes...)
	dirBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(dirBytes, m.Direction)
	buf = append(buf, dirBytes...)
	// 前车信息
	vn := []byte(m.VehicleNo)
	buf = append(buf, byte(len(vn)))
	buf = append(buf, vn...)
	vsBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(vsBytes, m.VehicleSpeed)
	buf = append(buf, vsBytes...)
	distBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(distBytes, m.Distance)
	buf = append(buf, distBytes...)
	buf = append(buf, m.DataItems...)
	return buf, nil
}

func (m *ADASDataMessage) Unmarshal(data []byte) error {
	if len(data) < 47 {
		return fmt.Errorf("1045 ADAS data too short")
	}
	m.VehicleColor = data[0]
	plate := data[1:22]
	for len(plate) > 0 && plate[len(plate)-1] == 0 {
		plate = plate[:len(plate)-1]
	}
	m.VehiclePlate = string(plate)
	t := make([]byte, 0, 12)
	for i := 0; i < 6; i++ {
		b := data[22+i]
		t = append(t, (b>>4)+'0', (b&0x0F)+'0')
	}
	m.AlarmTime = string(t)
	m.AlarmType = binary.BigEndian.Uint16(data[28:30])
	m.AlarmLevel = data[30]
	m.Lat = float64(binary.BigEndian.Uint32(data[31:35])) / 1000000.0
	m.Lon = float64(binary.BigEndian.Uint32(data[35:39])) / 1000000.0
	m.Speed = binary.BigEndian.Uint16(data[39:41])
	m.Direction = binary.BigEndian.Uint16(data[41:43])
	offset := 43
	if offset < len(data) {
		vnLen := int(data[offset])
		offset++
		if offset+vnLen <= len(data) {
			m.VehicleNo = string(data[offset : offset+vnLen])
			offset += vnLen
		}
	}
	if offset+4 <= len(data) {
		m.VehicleSpeed = binary.BigEndian.Uint16(data[offset : offset+2])
		m.Distance = binary.BigEndian.Uint16(data[offset+2 : offset+4])
		offset += 4
	}
	if offset < len(data) {
		m.DataItems = make([]byte, len(data)-offset)
		copy(m.DataItems, data[offset:])
	}
	return nil
}

// TireAlarmMessage 胎压报警
type TireAlarmMessage struct {
	VehicleColor byte
	VehiclePlate string
	AlarmTime    string
	TirePosition byte   // 胎位
	Pressure     uint16 // 胎压(kPa*10)
	Temp         int16  // 胎温(℃)
	Battery      byte   // 电池电量
	AlarmFlag    byte   // 报警标志
}

func (m *TireAlarmMessage) MsgID() uint16 { return MsgIDTireAlarm }

func (m *TireAlarmMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 0, 40)
	buf = append(buf, m.VehicleColor)
	plate := []byte(m.VehiclePlate)
	for len(plate) < 21 {
		plate = append(plate, 0)
	}
	buf = append(buf, plate[:21]...)
	for len(m.AlarmTime) < 12 {
		m.AlarmTime = "0" + m.AlarmTime
	}
	for i := 0; i < 6; i++ {
		buf = append(buf, ((m.AlarmTime[i*2]-'0')<<4)|(m.AlarmTime[i*2+1]-'0'))
	}
	buf = append(buf, m.TirePosition)
	pBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(pBytes, m.Pressure)
	buf = append(buf, pBytes...)
	tBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(tBytes, uint16(m.Temp))
	buf = append(buf, tBytes...)
	buf = append(buf, m.Battery)
	buf = append(buf, m.AlarmFlag)
	return buf, nil
}

func (m *TireAlarmMessage) Unmarshal(data []byte) error {
	// color(1) + plate(21) + time(6) + tirePos(1) + pressure(2) + temp(2) + battery(1) + alarmFlag(1) = 35
	if len(data) < 35 {
		return fmt.Errorf("1045 tire alarm too short: %d", len(data))
	}
	m.VehicleColor = data[0]
	plate := data[1:22]
	for len(plate) > 0 && plate[len(plate)-1] == 0 {
		plate = plate[:len(plate)-1]
	}
	m.VehiclePlate = string(plate)
	t := make([]byte, 0, 12)
	for i := 0; i < 6; i++ {
		b := data[22+i]
		t = append(t, (b>>4)+'0', (b&0x0F)+'0')
	}
	m.AlarmTime = string(t)
	m.TirePosition = data[28]
	m.Pressure = binary.BigEndian.Uint16(data[29:31])
	m.Temp = int16(binary.BigEndian.Uint16(data[31:33]))
	m.Battery = data[33]
	m.AlarmFlag = data[34]
	return nil
}
