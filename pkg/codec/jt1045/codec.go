package jt1045

import (
	"encoding/binary"
	"fmt"
	"math"
)

// ADASAlarmMessage ADAS报警消息
type ADASAlarmMessage struct {
	VehicleColor byte
	VehiclePlate string
	AlarmTime    string
	AlarmType    uint16
	AlarmLevel   byte
	AlarmParam   uint16
	Lat          float64
	Lon          float64
	Altitude     uint16
	Speed        uint16
	Direction    uint16
}

func (m *ADASAlarmMessage) MsgID() uint16 { return MsgIDADASAlarm }

func (m *ADASAlarmMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 0, 50)
	buf = append(buf, m.VehicleColor)
	plate := []byte(m.VehiclePlate)
	for len(plate) < 21 {
		plate = append(plate, 0)
	}
	buf = append(buf, plate[:21]...)

	// 时间BCD 6字节
	t := m.AlarmTime
	for len(t) < 12 {
		t = "0" + t
	}
	for i := 0; i < 6; i++ {
		high := t[i*2] - '0'
		low := t[i*2+1] - '0'
		buf = append(buf, (high<<4)|low)
	}

	typeBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(typeBytes, m.AlarmType)
	buf = append(buf, typeBytes...)
	buf = append(buf, m.AlarmLevel)

	paramBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(paramBytes, m.AlarmParam)
	buf = append(buf, paramBytes...)

	latBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(latBytes, uint32(math.Abs(m.Lat)*1000000))
	buf = append(buf, latBytes...)

	lonBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lonBytes, uint32(math.Abs(m.Lon)*1000000))
	buf = append(buf, lonBytes...)

	altBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(altBytes, m.Altitude)
	buf = append(buf, altBytes...)

	speedBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(speedBytes, m.Speed)
	buf = append(buf, speedBytes...)

	dirBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(dirBytes, m.Direction)
	buf = append(buf, dirBytes...)

	return buf, nil
}

func (m *ADASAlarmMessage) Unmarshal(data []byte) error {
	// 1(color)+21(plate)+6(time)+2(type)+1(level)+2(param)+4(lat)+4(lon)+2(alt)+2(speed)+2(dir) = 47
	if len(data) < 47 {
		return fmt.Errorf("1045 ADAS alarm too short: %d", len(data))
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
	m.AlarmParam = binary.BigEndian.Uint16(data[31:33])
	m.Lat = float64(binary.BigEndian.Uint32(data[33:37])) / 1000000.0
	m.Lon = float64(binary.BigEndian.Uint32(data[37:41])) / 1000000.0
	m.Altitude = binary.BigEndian.Uint16(data[41:43])
	m.Speed = binary.BigEndian.Uint16(data[43:45])
	m.Direction = binary.BigEndian.Uint16(data[45:47])
	return nil
}

// DSMAlarmMessage DSM报警（驾驶员状态监控）
type DSMAlarmMessage struct {
	VehicleColor byte
	VehiclePlate string
	AlarmTime    string
	AlarmType    uint16
	AlarmLevel   byte
	AlarmParam   uint16
	Lat          float64
	Lon          float64
	Speed        uint16
	Direction    uint16
}

func (m *DSMAlarmMessage) MsgID() uint16 { return MsgIDDSMAlarm }

func (m *DSMAlarmMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 0, 50)
	buf = append(buf, m.VehicleColor)
	plate := []byte(m.VehiclePlate)
	for len(plate) < 21 {
		plate = append(plate, 0)
	}
	buf = append(buf, plate[:21]...)

	t := m.AlarmTime
	for len(t) < 12 {
		t = "0" + t
	}
	for i := 0; i < 6; i++ {
		high := t[i*2] - '0'
		low := t[i*2+1] - '0'
		buf = append(buf, (high<<4)|low)
	}

	typeBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(typeBytes, m.AlarmType)
	buf = append(buf, typeBytes...)
	buf = append(buf, m.AlarmLevel)

	paramBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(paramBytes, m.AlarmParam)
	buf = append(buf, paramBytes...)

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
	return buf, nil
}

func (m *DSMAlarmMessage) Unmarshal(data []byte) error {
	// 1(color)+21(plate)+6(time)+2(type)+1(level)+2(param)+4(lat)+4(lon)+2(speed)+2(dir) = 45
	if len(data) < 45 {
		return fmt.Errorf("1045 DSM alarm too short: %d", len(data))
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
	m.AlarmParam = binary.BigEndian.Uint16(data[31:33])
	m.Lat = float64(binary.BigEndian.Uint32(data[33:37])) / 1000000.0
	m.Lon = float64(binary.BigEndian.Uint32(data[37:41])) / 1000000.0
	if len(data) >= 45 {
		m.Speed = binary.BigEndian.Uint16(data[41:43])
		m.Direction = binary.BigEndian.Uint16(data[43:45])
	}
	return nil
}

// MsgName 消息名称
func MsgName(msgID uint16) string {
	for _, m := range AllMessages1045 {
		if m.ID == msgID {
			return m.Name
		}
	}
	return fmt.Sprintf("未知消息(0x%04X)", msgID)
}
