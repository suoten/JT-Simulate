package jt905

import (
	"encoding/binary"
	"fmt"
	"math"
)

// TaxiOperateMessage 0x0B01 运营数据上报
type TaxiOperateMessage struct {
	VehicleColor byte
	VehiclePlate string
	OperateType  byte   // 1=上车, 2=下车
	GetOnTime    string // YYMMDDHHmmss
	GetOffTime   string
	GetOnLat     float64
	GetOnLon     float64
	GetOffLat    float64
	GetOffLon    float64
	Distance     uint32 // 行驶距离（米）
	Amount       uint32 // 金额（分）
}

func (m *TaxiOperateMessage) MsgID() uint16 { return MsgIDTaxiOperate }

func (m *TaxiOperateMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 0, 60)
	buf = append(buf, m.VehicleColor)
	plate := []byte(m.VehiclePlate)
	for len(plate) < 21 {
		plate = append(plate, 0)
	}
	buf = append(buf, plate[:21]...)
	buf = append(buf, m.OperateType)

	for _, timeStr := range []string{m.GetOnTime, m.GetOffTime} {
		for len(timeStr) < 12 {
			timeStr = "0" + timeStr
		}
		for i := 0; i < 6; i++ {
			high := timeStr[i*2] - '0'
			low := timeStr[i*2+1] - '0'
			buf = append(buf, (high<<4)|low)
		}
	}

	latBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(latBytes, uint32(math.Abs(m.GetOnLat)*1000000))
	buf = append(buf, latBytes...)

	lonBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lonBytes, uint32(math.Abs(m.GetOnLon)*1000000))
	buf = append(buf, lonBytes...)

	latOffBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(latOffBytes, uint32(math.Abs(m.GetOffLat)*1000000))
	buf = append(buf, latOffBytes...)

	lonOffBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lonOffBytes, uint32(math.Abs(m.GetOffLon)*1000000))
	buf = append(buf, lonOffBytes...)

	distBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(distBytes, m.Distance)
	buf = append(buf, distBytes...)

	amtBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(amtBytes, m.Amount)
	buf = append(buf, amtBytes...)

	return buf, nil
}

func (m *TaxiOperateMessage) Unmarshal(data []byte) error {
	if len(data) < 50 {
		return fmt.Errorf("905 operate data too short")
	}
	m.VehicleColor = data[0]
	plate := data[1:22]
	for len(plate) > 0 && plate[len(plate)-1] == 0 {
		plate = plate[:len(plate)-1]
	}
	m.VehiclePlate = string(plate)
	m.OperateType = data[22]

	decodeBCD := func(offset int) string {
		s := make([]byte, 0, 12)
		for i := 0; i < 6; i++ {
			b := data[offset+i]
			s = append(s, (b>>4)+'0', (b&0x0F)+'0')
		}
		return string(s)
	}
	m.GetOnTime = decodeBCD(23)
	m.GetOffTime = decodeBCD(29)
	m.GetOnLat = float64(binary.BigEndian.Uint32(data[35:39])) / 1000000.0
	m.GetOnLon = float64(binary.BigEndian.Uint32(data[39:43])) / 1000000.0
	m.GetOffLat = float64(binary.BigEndian.Uint32(data[43:47])) / 1000000.0
	m.GetOffLon = float64(binary.BigEndian.Uint32(data[47:51])) / 1000000.0
	if len(data) >= 59 {
		m.Distance = binary.BigEndian.Uint32(data[51:55])
		m.Amount = binary.BigEndian.Uint32(data[55:59])
	}
	return nil
}

// MsgName 消息名称
func MsgName(msgID uint16) string {
	for _, m := range AllMessages905 {
		if m.ID == msgID {
			return m.Name
		}
	}
	return fmt.Sprintf("未知消息(0x%04X)", msgID)
}
