package jt1253

import (
	"encoding/binary"
	"fmt"
	"math"
)

// ==================== JT/T 1253-2019 完整消息ID ====================

const (
	MsgIDEWaybill        uint16 = 0x0D01 // 电子运单上报
	MsgIDHazardAlarm     uint16 = 0x0D02 // 危险品报警
	MsgIDHazardStatus    uint16 = 0x0D03 // 危险品状态
	MsgIDUnloadReport    uint16 = 0x0D04 // 卸载报告
	MsgIDHazardMonitor   uint16 = 0x0D05 // 危险品监测数据
	MsgIDTempHumidity    uint16 = 0x0D06 // 温湿度数据
	MsgIDPressure         uint16 = 0x0D07 // 压力数据
	MsgIDLevel           uint16 = 0x0D08 // 液位数据
	MsgIDHazardRoute     uint16 = 0x0D09 // 危险品路线偏离
	MsgIDHazardParking   uint16 = 0x0D0A // 危险品停车超时
	MsgIDHazardSpeed     uint16 = 0x0D0B // 危险品超速
	MsgIDEWaybillQuery   uint16 = 0x8D01 // 电子运单查询
	MsgIDEWaybillResp    uint16 = 0x8D02 // 电子运单下发
	MsgIDHazardAlarmAck  uint16 = 0x8D03 // 危险品报警确认
	MsgIDHazardCmd       uint16 = 0x8D04 // 危险品控制指令
	MsgIDRouteSet        uint16 = 0x8D05 // 危险品路线设置
	MsgIDRouteDel        uint16 = 0x8D06 // 危险品路线删除
	MsgIDParkingSet      uint16 = 0x8D07 // 停车区域设置
	MsgIDParkingDel      uint16 = 0x8D08 // 停车区域删除
	MsgIDSpeedSet        uint16 = 0x8D09 // 限速设置
	MsgIDTempThreshold   uint16 = 0x8D0A // 温湿度阈值设置
	MsgIDPressureThreshold uint16 = 0x8D0B // 压力阈值设置
	MsgIDLevelThreshold  uint16 = 0x8D0C // 液位阈值设置
)

// MsgInfo1253 1253消息信息
type MsgInfo1253 struct {
	ID   uint16
	Name string
	Dir  string
}

// AllMessages1253 1253协议全部消息列表
var AllMessages1253 = []MsgInfo1253{
	{0x0D01, "电子运单上报", "up"},
	{0x0D02, "危险品报警", "up"},
	{0x0D03, "危险品状态", "up"},
	{0x0D04, "卸载报告", "up"},
	{0x0D05, "危险品监测数据", "up"},
	{0x0D06, "温湿度数据", "up"},
	{0x0D07, "压力数据", "up"},
	{0x0D08, "液位数据", "up"},
	{0x0D09, "危险品路线偏离", "up"},
	{0x0D0A, "危险品停车超时", "up"},
	{0x0D0B, "危险品超速", "up"},
	{0x8D01, "电子运单查询", "down"},
	{0x8D02, "电子运单下发", "down"},
	{0x8D03, "危险品报警确认", "down"},
	{0x8D04, "危险品控制指令", "down"},
	{0x8D05, "危险品路线设置", "down"},
	{0x8D06, "危险品路线删除", "down"},
	{0x8D07, "停车区域设置", "down"},
	{0x8D08, "停车区域删除", "down"},
	{0x8D09, "限速设置", "down"},
	{0x8D0A, "温湿度阈值设置", "down"},
	{0x8D0B, "压力阈值设置", "down"},
	{0x8D0C, "液位阈值设置", "down"},
}

// HazardStatusMessage 危险品状态
type HazardStatusMessage struct {
	VehicleColor byte
	VehiclePlate string
	StatusTime   string
	HazardClass  byte
	HazardState  byte   // 状态标志
	Lat          float64
	Lon          float64
	Speed        uint16
	Direction    uint16
	Temp         int16  // 温度(℃)
	Humidity     uint16 // 湿度(%)
	Pressure     uint32 // 压力
	Level        uint16 // 液位
}

func (m *HazardStatusMessage) MsgID() uint16 { return MsgIDHazardStatus }

func (m *HazardStatusMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 0, 55)
	buf = append(buf, m.VehicleColor)
	plate := []byte(m.VehiclePlate)
	for len(plate) < 21 {
		plate = append(plate, 0)
	}
	buf = append(buf, plate[:21]...)
	for len(m.StatusTime) < 12 {
		m.StatusTime = "0" + m.StatusTime
	}
	for i := 0; i < 6; i++ {
		buf = append(buf, ((m.StatusTime[i*2]-'0')<<4)|(m.StatusTime[i*2+1]-'0'))
	}
	buf = append(buf, m.HazardClass)
	buf = append(buf, m.HazardState)
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
	tempBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(tempBytes, uint16(m.Temp))
	buf = append(buf, tempBytes...)
	humBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(humBytes, m.Humidity)
	buf = append(buf, humBytes...)
	pressBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(pressBytes, m.Pressure)
	buf = append(buf, pressBytes...)
	levelBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(levelBytes, m.Level)
	buf = append(buf, levelBytes...)
	return buf, nil
}

func (m *HazardStatusMessage) Unmarshal(data []byte) error {
	if len(data) < 50 {
		return fmt.Errorf("1253 hazard status too short")
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
	m.StatusTime = string(t)
	m.HazardClass = data[28]
	m.HazardState = data[29]
	m.Lat = float64(binary.BigEndian.Uint32(data[30:34])) / 1000000.0
	m.Lon = float64(binary.BigEndian.Uint32(data[34:38])) / 1000000.0
	m.Speed = binary.BigEndian.Uint16(data[38:40])
	m.Direction = binary.BigEndian.Uint16(data[40:42])
	m.Temp = int16(binary.BigEndian.Uint16(data[42:44]))
	m.Humidity = binary.BigEndian.Uint16(data[44:46])
	m.Pressure = binary.BigEndian.Uint32(data[46:50])
	if len(data) >= 52 {
		m.Level = binary.BigEndian.Uint16(data[50:52])
	}
	return nil
}

// UnloadReportMessage 卸载报告
type UnloadReportMessage struct {
	VehicleColor byte
	VehiclePlate string
	UnloadTime   string
	HazardClass  byte
	UnloadAmount uint32 // 卸载数量
	UnloadPlace  string
	Lat          float64
	Lon          float64
}

func (m *UnloadReportMessage) MsgID() uint16 { return MsgIDUnloadReport }

func (m *UnloadReportMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 0, 60)
	buf = append(buf, m.VehicleColor)
	plate := []byte(m.VehiclePlate)
	for len(plate) < 21 {
		plate = append(plate, 0)
	}
	buf = append(buf, plate[:21]...)
	for len(m.UnloadTime) < 12 {
		m.UnloadTime = "0" + m.UnloadTime
	}
	for i := 0; i < 6; i++ {
		buf = append(buf, ((m.UnloadTime[i*2]-'0')<<4)|(m.UnloadTime[i*2+1]-'0'))
	}
	buf = append(buf, m.HazardClass)
	amtBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(amtBytes, m.UnloadAmount)
	buf = append(buf, amtBytes...)
	place := []byte(m.UnloadPlace)
	buf = append(buf, byte(len(place)))
	buf = append(buf, place...)
	latBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(latBytes, uint32(math.Abs(m.Lat)*1000000))
	buf = append(buf, latBytes...)
	lonBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lonBytes, uint32(math.Abs(m.Lon)*1000000))
	buf = append(buf, lonBytes...)
	return buf, nil
}

func (m *UnloadReportMessage) Unmarshal(data []byte) error {
	if len(data) < 40 {
		return fmt.Errorf("1253 unload report too short")
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
	m.UnloadTime = string(t)
	m.HazardClass = data[28]
	m.UnloadAmount = binary.BigEndian.Uint32(data[29:33])
	placeLen := int(data[33])
	offset := 34
	if offset+placeLen <= len(data) {
		m.UnloadPlace = string(data[offset : offset+placeLen])
		offset += placeLen
	}
	// lat(4) + lon(4)
	if offset+8 <= len(data) {
		m.Lat = float64(binary.BigEndian.Uint32(data[offset:offset+4])) / 1000000.0
		m.Lon = float64(binary.BigEndian.Uint32(data[offset+4:offset+8])) / 1000000.0
	}
	return nil
}
