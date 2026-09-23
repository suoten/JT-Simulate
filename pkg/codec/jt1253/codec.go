package jt1253

import (
	"encoding/binary"
	"fmt"
	"math"
)

// EWaybillMessage 电子运单
type EWaybillMessage struct {
	VehicleColor  byte
	VehiclePlate  string
	WaybillID     string
	HazardClass   byte
	HazardName    string
	Weight        uint32
	Origin        string
	Destination   string
	LoadTime      string
	UnloadTime    string
}

func (m *EWaybillMessage) MsgID() uint16 { return MsgIDEWaybill }

func (m *EWaybillMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 0, 120)
	buf = append(buf, m.VehicleColor)
	plate := []byte(m.VehiclePlate)
	for len(plate) < 21 {
		plate = append(plate, 0)
	}
	buf = append(buf, plate[:21]...)

	wbID := []byte(m.WaybillID)
	for len(wbID) < 20 {
		wbID = append(wbID, 0)
	}
	buf = append(buf, wbID[:20]...)
	buf = append(buf, m.HazardClass)

	hn := []byte(m.HazardName)
	buf = append(buf, byte(len(hn)))
	buf = append(buf, hn...)

	weightBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(weightBytes, m.Weight)
	buf = append(buf, weightBytes...)

	origin := []byte(m.Origin)
	buf = append(buf, byte(len(origin)))
	buf = append(buf, origin...)

	dest := []byte(m.Destination)
	buf = append(buf, byte(len(dest)))
	buf = append(buf, dest...)

	// 时间BCD
	for _, timeStr := range []string{m.LoadTime, m.UnloadTime} {
		for len(timeStr) < 12 {
			timeStr = "0" + timeStr
		}
		for i := 0; i < 6; i++ {
			high := timeStr[i*2] - '0'
			low := timeStr[i*2+1] - '0'
			buf = append(buf, (high<<4)|low)
		}
	}

	return buf, nil
}

func (m *EWaybillMessage) Unmarshal(data []byte) error {
	if len(data) < 24 {
		return fmt.Errorf("1253 ewaybill too short")
	}
	m.VehicleColor = data[0]
	plate := data[1:22]
	for len(plate) > 0 && plate[len(plate)-1] == 0 {
		plate = plate[:len(plate)-1]
	}
	m.VehiclePlate = string(plate)
	wbID := data[22:42]
	for len(wbID) > 0 && wbID[len(wbID)-1] == 0 {
		wbID = wbID[:len(wbID)-1]
	}
	m.WaybillID = string(wbID)
	m.HazardClass = data[42]
	offset := 43
	// hazardName (1+N)
	if offset >= len(data) {
		return nil
	}
	hnLen := int(data[offset])
	offset++
	if offset+hnLen > len(data) {
		return nil
	}
	m.HazardName = string(data[offset : offset+hnLen])
	offset += hnLen
	// weight 4 bytes
	if offset+4 > len(data) {
		return nil
	}
	m.Weight = binary.BigEndian.Uint32(data[offset : offset+4])
	offset += 4
	// origin (1+N)
	if offset >= len(data) {
		return nil
	}
	origLen := int(data[offset])
	offset++
	if offset+origLen > len(data) {
		return nil
	}
	m.Origin = string(data[offset : offset+origLen])
	offset += origLen
	// destination (1+N)
	if offset >= len(data) {
		return nil
	}
	destLen := int(data[offset])
	offset++
	if offset+destLen > len(data) {
		return nil
	}
	m.Destination = string(data[offset : offset+destLen])
	offset += destLen
	// loadTime BCD 6 bytes
	if offset+6 > len(data) {
		return nil
	}
	t1 := make([]byte, 0, 12)
	for i := 0; i < 6; i++ {
		b := data[offset+i]
		t1 = append(t1, (b>>4)+'0', (b&0x0F)+'0')
	}
	m.LoadTime = string(t1)
	offset += 6
	// unloadTime BCD 6 bytes
	if offset+6 > len(data) {
		return nil
	}
	t2 := make([]byte, 0, 12)
	for i := 0; i < 6; i++ {
		b := data[offset+i]
		t2 = append(t2, (b>>4)+'0', (b&0x0F)+'0')
	}
	m.UnloadTime = string(t2)
	return nil
}

// HazardAlarmMessage 危险品报警
type HazardAlarmMessage struct {
	VehicleColor byte
	VehiclePlate string
	AlarmType    uint16
	AlarmLevel   byte
	AlarmTime    string
	Lat          float64
	Lon          float64
	Speed        uint16
	Direction    uint16
}

func (m *HazardAlarmMessage) MsgID() uint16 { return MsgIDHazardAlarm }

func (m *HazardAlarmMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 0, 50)
	buf = append(buf, m.VehicleColor)
	plate := []byte(m.VehiclePlate)
	for len(plate) < 21 {
		plate = append(plate, 0)
	}
	buf = append(buf, plate[:21]...)

	typeBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(typeBytes, m.AlarmType)
	buf = append(buf, typeBytes...)
	buf = append(buf, m.AlarmLevel)

	t := m.AlarmTime
	for len(t) < 12 {
		t = "0" + t
	}
	for i := 0; i < 6; i++ {
		high := t[i*2] - '0'
		low := t[i*2+1] - '0'
		buf = append(buf, (high<<4)|low)
	}

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

func (m *HazardAlarmMessage) Unmarshal(data []byte) error {
	// 1(color)+21(plate)+2(type)+1(level)+6(time)+4(lat)+4(lon)+2(speed)+2(dir) = 43
	if len(data) < 43 {
		return fmt.Errorf("1253 hazard alarm too short: %d", len(data))
	}
	m.VehicleColor = data[0]
	plate := data[1:22]
	for len(plate) > 0 && plate[len(plate)-1] == 0 {
		plate = plate[:len(plate)-1]
	}
	m.VehiclePlate = string(plate)
	m.AlarmType = binary.BigEndian.Uint16(data[22:24])
	m.AlarmLevel = data[24]
	// BCD time 6 bytes
	t := make([]byte, 0, 12)
	for i := 0; i < 6; i++ {
		b := data[25+i]
		t = append(t, (b>>4)+'0', (b&0x0F)+'0')
	}
	m.AlarmTime = string(t)
	m.Lat = float64(binary.BigEndian.Uint32(data[31:35])) / 1000000.0
	m.Lon = float64(binary.BigEndian.Uint32(data[35:39])) / 1000000.0
	m.Speed = binary.BigEndian.Uint16(data[39:41])
	m.Direction = binary.BigEndian.Uint16(data[41:43])
	return nil
}

// MsgName 消息名称
func MsgName(msgID uint16) string {
	for _, m := range AllMessages1253 {
		if m.ID == msgID {
			return m.Name
		}
	}
	return fmt.Sprintf("未知消息(0x%04X)", msgID)
}
