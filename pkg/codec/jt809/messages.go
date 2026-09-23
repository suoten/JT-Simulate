package jt809

import (
	"encoding/binary"
	"fmt"
	"math"
)

// ConnectReqMessage 0x1001 下级平台请求连接
type ConnectReqMessage struct {
	UserName   uint32
	Password   string
	DownLinkIP string
	DownLinkPort uint32
}

func (m *ConnectReqMessage) MsgID() uint16 { return MsgIDUpConnectReq }

func (m *ConnectReqMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 0, 30)
	userBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(userBytes, m.UserName)
	buf = append(buf, userBytes...)

	pwd := []byte(m.Password)
	for len(pwd) < 10 {
		pwd = append(pwd, 0)
	}
	buf = append(buf, pwd[:10]...)

	ip := []byte(m.DownLinkIP)
	for len(ip) < 10 {
		ip = append(ip, 0)
	}
	buf = append(buf, ip[:10]...)

	portBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(portBytes, m.DownLinkPort)
	buf = append(buf, portBytes...)

	return buf, nil
}

func (m *ConnectReqMessage) Unmarshal(data []byte) error {
	if len(data) < 28 {
		return fmt.Errorf("809 connect req too short: %d", len(data))
	}
	m.UserName = binary.BigEndian.Uint32(data[0:4])
	pwd := data[4:14]
	for len(pwd) > 0 && pwd[len(pwd)-1] == 0 {
		pwd = pwd[:len(pwd)-1]
	}
	m.Password = string(pwd)
	ip := data[14:24]
	for len(ip) > 0 && ip[len(ip)-1] == 0 {
		ip = ip[:len(ip)-1]
	}
	m.DownLinkIP = string(ip)
	m.DownLinkPort = binary.BigEndian.Uint32(data[24:28])
	return nil
}

// ConnectRspMessage 0x1002 连接应答
type ConnectRspMessage struct {
	Result      byte
	VerifyCode  uint32
	DownLinkIP  string
	DownLinkPort uint32
}

func (m *ConnectRspMessage) MsgID() uint16 { return MsgIDUpConnectRsp }

func (m *ConnectRspMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 0, 20)
	buf = append(buf, m.Result)
	vcBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(vcBytes, m.VerifyCode)
	buf = append(buf, vcBytes...)
	ip := []byte(m.DownLinkIP)
	for len(ip) < 10 {
		ip = append(ip, 0)
	}
	buf = append(buf, ip[:10]...)
	portBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(portBytes, m.DownLinkPort)
	buf = append(buf, portBytes...)
	return buf, nil
}

func (m *ConnectRspMessage) Unmarshal(data []byte) error {
	if len(data) < 19 {
		return fmt.Errorf("809 connect rsp too short")
	}
	m.Result = data[0]
	m.VerifyCode = binary.BigEndian.Uint32(data[1:5])
	ip := data[5:15]
	for len(ip) > 0 && ip[len(ip)-1] == 0 {
		ip = ip[:len(ip)-1]
	}
	m.DownLinkIP = string(ip)
	m.DownLinkPort = binary.BigEndian.Uint32(data[15:19])
	return nil
}

// LinktestReqMessage 0x1005 链路保持请求
type LinktestReqMessage struct{}

func (m *LinktestReqMessage) MsgID() uint16                  { return MsgIDUpLinktestReq }
func (m *LinktestReqMessage) Marshal() ([]byte, error)       { return nil, nil }
func (m *LinktestReqMessage) Unmarshal(data []byte) error     { return nil }

// LinktestRspMessage 0x1006 链路保持应答
type LinktestRspMessage struct{}

func (m *LinktestRspMessage) MsgID() uint16                  { return MsgIDUpLinktestRsp }
func (m *LinktestRspMessage) Marshal() ([]byte, error)       { return nil, nil }
func (m *LinktestRspMessage) Unmarshal(data []byte) error     { return nil }

// LocationMessage 0x1205 上报车辆定位
type LocationMessage struct {
	VehicleColor byte
	VehiclePlate string
	AlarmFlag    uint32
	Lat          float64
	Lon          float64
	Speed        uint16
	Direction    uint16
	Time         string
}

func (m *LocationMessage) MsgID() uint16 { return MsgIDLocationMsg }

func (m *LocationMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 0, 40)
	buf = append(buf, m.VehicleColor)
	plate := []byte(m.VehiclePlate)
	for len(plate) < 21 {
		plate = append(plate, 0)
	}
	buf = append(buf, plate[:21]...)

	alarm := make([]byte, 4)
	binary.BigEndian.PutUint32(alarm, m.AlarmFlag)
	buf = append(buf, alarm...)

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

	// 时间BCD 6字节
	t := []byte(m.Time)
	for len(t) < 12 {
		t = append([]byte{'0'}, t...)
	}
	for i := 0; i < 6; i++ {
		high := t[i*2] - '0'
		low := t[i*2+1] - '0'
		buf = append(buf, (high<<4)|low)
	}

	return buf, nil
}

func (m *LocationMessage) Unmarshal(data []byte) error {
	if len(data) < 38 {
		return fmt.Errorf("809 location too short")
	}
	m.VehicleColor = data[0]
	plate := data[1:22]
	for len(plate) > 0 && plate[len(plate)-1] == 0 {
		plate = plate[:len(plate)-1]
	}
	m.VehiclePlate = string(plate)
	m.AlarmFlag = binary.BigEndian.Uint32(data[22:26])
	m.Lat = float64(binary.BigEndian.Uint32(data[26:30])) / 1000000.0
	m.Lon = float64(binary.BigEndian.Uint32(data[30:34])) / 1000000.0
	m.Speed = binary.BigEndian.Uint16(data[34:36])
	m.Direction = binary.BigEndian.Uint16(data[36:38])
	// 时间 BCD 6字节
	if len(data) >= 44 {
		t := make([]byte, 0, 12)
		for i := 0; i < 6; i++ {
			b := data[38+i]
			t = append(t, (b>>4)+'0', (b&0x0F)+'0')
		}
		m.Time = string(t)
	}
	return nil
}

// AlarmMessage 0x1300 上报报警
type AlarmMessage struct {
	VehicleColor  byte
	VehiclePlate  string
	AlarmSource   byte
	AlarmType     uint16
	AlarmTime     string
	Lat           float64
	Lon           float64
	Speed         uint16
	Direction     uint16
}

func (m *AlarmMessage) MsgID() uint16 { return MsgIDAlarmMsg }

func (m *AlarmMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 0, 45)
	buf = append(buf, m.VehicleColor)
	plate := []byte(m.VehiclePlate)
	for len(plate) < 21 {
		plate = append(plate, 0)
	}
	buf = append(buf, plate[:21]...)
	buf = append(buf, m.AlarmSource)
	atBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(atBytes, m.AlarmType)
	buf = append(buf, atBytes...)

	// 时间BCD 6字节
	t := []byte(m.AlarmTime)
	for len(t) < 12 {
		t = append([]byte{'0'}, t...)
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

func (m *AlarmMessage) Unmarshal(data []byte) error {
	// color(1) + plate(21) + alarmSource(1) + alarmType(2) + timeBCD(6) + lat(4) + lon(4) + speed(2) + dir(2) = 43
	if len(data) < 43 {
		return fmt.Errorf("809 alarm too short: %d", len(data))
	}
	m.VehicleColor = data[0]
	plate := data[1:22]
	for len(plate) > 0 && plate[len(plate)-1] == 0 {
		plate = plate[:len(plate)-1]
	}
	m.VehiclePlate = string(plate)
	m.AlarmSource = data[22]
	m.AlarmType = binary.BigEndian.Uint16(data[23:25])
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
