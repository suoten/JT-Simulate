package jt809

import (
	"encoding/binary"
	"fmt"
	"math"
)

// StartupMessage 0x1202 车辆启动
type StartupMessage struct {
	VehicleColor byte
	VehiclePlate string
	StartupTime  string // YYMMDDHHmmss
	Lat          float64
	Lon          float64
	Speed        uint16
	Direction    uint16
}

func (m *StartupMessage) MsgID() uint16 { return MsgIDStartupMsg }

func (m *StartupMessage) Marshal() ([]byte, error) {
	return encodeVehicleLocBody(m.VehicleColor, m.VehiclePlate, m.StartupTime, m.Lat, m.Lon, m.Speed, m.Direction), nil
}

func (m *StartupMessage) Unmarshal(data []byte) error {
	c, p, t, lat, lon, spd, dir, err := decodeVehicleLocBody(data)
	if err != nil {
		return err
	}
	m.VehicleColor = c
	m.VehiclePlate = p
	m.StartupTime = t
	m.Lat = lat
	m.Lon = lon
	m.Speed = spd
	m.Direction = dir
	return nil
}

// ShutdownMessage 0x1203 车辆关闭
type ShutdownMessage struct {
	VehicleColor  byte
	VehiclePlate  string
	ShutdownTime  string
	Lat           float64
	Lon           float64
	Speed         uint16
	Direction     uint16
}

func (m *ShutdownMessage) MsgID() uint16 { return MsgIDShutdownMsg }

func (m *ShutdownMessage) Marshal() ([]byte, error) {
	return encodeVehicleLocBody(m.VehicleColor, m.VehiclePlate, m.ShutdownTime, m.Lat, m.Lon, m.Speed, m.Direction), nil
}

func (m *ShutdownMessage) Unmarshal(data []byte) error {
	c, p, t, lat, lon, spd, dir, err := decodeVehicleLocBody(data)
	if err != nil {
		return err
	}
	m.VehicleColor = c
	m.VehiclePlate = p
	m.ShutdownTime = t
	m.Lat = lat
	m.Lon = lon
	m.Speed = spd
	m.Direction = dir
	return nil
}

// VehicleInfoMessage 0x1200 车辆信息上报
type VehicleInfoMessage struct {
	VehicleColor   byte
	VehiclePlate   string
	VehicleNationalID string // 车辆国家编号(12字节)
	VIN            string // 车辆识别码(17字节)
	VehicleType    uint16 // 车辆类型
	FirstRegisterDate string // 首次注册日期 YYYYMMDD
	EngineNumber   string // 发动机号
	VehicleBrand   string // 车辆品牌
	VehicleModel   string // 车辆型号
}

func (m *VehicleInfoMessage) MsgID() uint16 { return MsgIDUploadCarMsg }

func (m *VehicleInfoMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 0, 80)
	buf = append(buf, m.VehicleColor)
	plate := []byte(m.VehiclePlate)
	for len(plate) < 21 {
		plate = append(plate, 0)
	}
	buf = append(buf, plate[:21]...)

	// 车辆国家编号 12字节
	natID := []byte(m.VehicleNationalID)
	for len(natID) < 12 {
		natID = append(natID, 0)
	}
	buf = append(buf, natID[:12]...)

	// VIN 17字节
	vin := []byte(m.VIN)
	for len(vin) < 17 {
		vin = append(vin, 0)
	}
	buf = append(buf, vin[:17]...)

	// 车辆类型 2字节
	typeBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(typeBytes, m.VehicleType)
	buf = append(buf, typeBytes...)

	// 首次注册日期 8字节
	date := []byte(m.FirstRegisterDate)
	for len(date) < 8 {
		date = append(date, 0)
	}
	buf = append(buf, date[:8]...)

	// 发动机号 变长(1+N)
	eng := []byte(m.EngineNumber)
	buf = append(buf, byte(len(eng)))
	buf = append(buf, eng...)

	// 车辆品牌 变长(1+N)
	brand := []byte(m.VehicleBrand)
	buf = append(buf, byte(len(brand)))
	buf = append(buf, brand...)

	// 车辆型号 变长(1+N)
	model := []byte(m.VehicleModel)
	buf = append(buf, byte(len(model)))
	buf = append(buf, model...)

	return buf, nil
}

func (m *VehicleInfoMessage) Unmarshal(data []byte) error {
	if len(data) < 62 {
		return fmt.Errorf("809 vehicle info too short: %d", len(data))
	}
	m.VehicleColor = data[0]
	m.VehiclePlate = trimNull809(data[1:22])
	m.VehicleNationalID = trimNull809(data[22:34])
	m.VIN = trimNull809(data[34:51])
	m.VehicleType = binary.BigEndian.Uint16(data[51:53])
	m.FirstRegisterDate = trimNull809(data[53:61])

	offset := 61
	if offset >= len(data) {
		return nil
	}
	engLen := int(data[offset])
	offset++
	if offset+engLen > len(data) {
		return nil
	}
	m.EngineNumber = string(data[offset : offset+engLen])
	offset += engLen

	if offset >= len(data) {
		return nil
	}
	brandLen := int(data[offset])
	offset++
	if offset+brandLen > len(data) {
		return nil
	}
	m.VehicleBrand = string(data[offset : offset+brandLen])
	offset += brandLen

	if offset >= len(data) {
		return nil
	}
	modelLen := int(data[offset])
	offset++
	if offset+modelLen > len(data) {
		return nil
	}
	m.VehicleModel = string(data[offset : offset+modelLen])
	return nil
}

// DisconnectReqMessage 0x1003 断开连接请求
type DisconnectReqMessage struct {
	UserName uint32
	Password string
}

func (m *DisconnectReqMessage) MsgID() uint16 { return MsgIDUpDisconnectReq }

func (m *DisconnectReqMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 0, 14)
	userBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(userBytes, m.UserName)
	buf = append(buf, userBytes...)
	pwd := []byte(m.Password)
	for len(pwd) < 10 {
		pwd = append(pwd, 0)
	}
	buf = append(buf, pwd[:10]...)
	return buf, nil
}

func (m *DisconnectReqMessage) Unmarshal(data []byte) error {
	if len(data) < 14 {
		return fmt.Errorf("809 disconnect req too short")
	}
	m.UserName = binary.BigEndian.Uint32(data[0:4])
	m.Password = trimNull809(data[4:14])
	return nil
}

// DisconnectRspMessage 0x9004 从链路断开应答
type DisconnectRspMessage struct {
	Result byte
}

func (m *DisconnectRspMessage) MsgID() uint16 { return MsgIDDnDisconnectRsp }

func (m *DisconnectRspMessage) Marshal() ([]byte, error) {
	return []byte{m.Result}, nil
}

func (m *DisconnectRspMessage) Unmarshal(data []byte) error {
	if len(data) < 1 {
		return fmt.Errorf("809 disconnect rsp too short")
	}
	m.Result = data[0]
	return nil
}

// DnLinktestReqMessage 0x9005 从链路连接保持请求
type DnLinktestReqMessage struct{}

func (m *DnLinktestReqMessage) MsgID() uint16                   { return MsgIDDnLinktestReq }
func (m *DnLinktestReqMessage) Marshal() ([]byte, error)        { return nil, nil }
func (m *DnLinktestReqMessage) Unmarshal(data []byte) error      { return nil }

// DnLinktestRspMessage 0x9006 从链路连接保持应答
type DnLinktestRspMessage struct{}

func (m *DnLinktestRspMessage) MsgID() uint16                   { return MsgIDDnLinktestRsp }
func (m *DnLinktestRspMessage) Marshal() ([]byte, error)        { return nil, nil }
func (m *DnLinktestRspMessage) Unmarshal(data []byte) error      { return nil }

// DriverInfoUpMessage 0x1701 驾驶员信息上报
type DriverInfoUpMessage struct {
	VehicleColor  byte
	VehiclePlate  string
	DriverName    string
	DriverID      string
	Licence       string
	OrgName       string
	UploadTime    string
}

func (m *DriverInfoUpMessage) MsgID() uint16 { return MsgIDDriverInfoMsg }

func (m *DriverInfoUpMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 0, 80)
	buf = append(buf, m.VehicleColor)
	plate := []byte(m.VehiclePlate)
	for len(plate) < 21 {
		plate = append(plate, 0)
	}
	buf = append(buf, plate[:21]...)

	name := []byte(m.DriverName)
	buf = append(buf, byte(len(name)))
	buf = append(buf, name...)

	id := []byte(m.DriverID)
	for len(id) < 20 {
		id = append(id, 0)
	}
	buf = append(buf, id[:20]...)

	lic := []byte(m.Licence)
	for len(lic) < 40 {
		lic = append(lic, 0)
	}
	buf = append(buf, lic[:40]...)

	org := []byte(m.OrgName)
	buf = append(buf, byte(len(org)))
	buf = append(buf, org...)

	t := m.UploadTime
	for len(t) < 12 {
		t = "0" + t
	}
	for i := 0; i < 6; i++ {
		buf = append(buf, ((t[i*2]-'0')<<4)|(t[i*2+1]-'0'))
	}
	return buf, nil
}

func (m *DriverInfoUpMessage) Unmarshal(data []byte) error {
	if len(data) < 23 {
		return fmt.Errorf("809 driver info too short")
	}
	m.VehicleColor = data[0]
	m.VehiclePlate = trimNull809(data[1:22])
	offset := 22
	if offset >= len(data) { return nil }
	nameLen := int(data[offset]); offset++
	if offset+nameLen > len(data) { return nil }
	m.DriverName = string(data[offset:offset+nameLen]); offset += nameLen
	if offset+20 > len(data) { return nil }
	m.DriverID = trimNull809(data[offset:offset+20]); offset += 20
	if offset+40 > len(data) { return nil }
	m.Licence = trimNull809(data[offset:offset+40]); offset += 40
	if offset >= len(data) { return nil }
	orgLen := int(data[offset]); offset++
	if offset+orgLen > len(data) { return nil }
	m.OrgName = string(data[offset:offset+orgLen]); offset += orgLen
	if offset+6 > len(data) { return nil }
	t := make([]byte, 0, 12)
	for i := 0; i < 6; i++ {
		b := data[offset+i]
		t = append(t, (b>>4)+'0', (b&0x0F)+'0')
	}
	m.UploadTime = string(t)
	return nil
}

// EWaybillUpMessage 0x1801 电子运单上报
type EWaybillUpMessage struct {
	VehicleColor byte
	VehiclePlate string
	WaybillID    string
	HazardClass  byte
	HazardName   string
	Weight       uint32
	Origin       string
	Destination  string
	LoadTime     string
	UnloadTime   string
}

func (m *EWaybillUpMessage) MsgID() uint16 { return MsgIDEWaybillUpMsg }

func (m *EWaybillUpMessage) Marshal() ([]byte, error) {
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

	for _, timeStr := range []string{m.LoadTime, m.UnloadTime} {
		for len(timeStr) < 12 {
			timeStr = "0" + timeStr
		}
		for i := 0; i < 6; i++ {
			buf = append(buf, ((timeStr[i*2]-'0')<<4)|(timeStr[i*2+1]-'0'))
		}
	}
	return buf, nil
}

func (m *EWaybillUpMessage) Unmarshal(data []byte) error {
	if len(data) < 24 {
		return fmt.Errorf("809 ewaybill too short")
	}
	m.VehicleColor = data[0]
	m.VehiclePlate = trimNull809(data[1:22])
	m.WaybillID = trimNull809(data[22:42])
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

// StatisticsMessage 0x1501 报文统计上报
type StatisticsMessage struct {
	VehicleColor   byte
	VehiclePlate   string
	TotalMsgCount  uint32 // 报文总数
	LocationCount  uint32 // 定位报文数
	AlarmCount     uint32 // 报警报文数
	OnlineDuration uint32 // 在线时长(秒)
	StartTime      string // 统计起始时间
	EndTime        string // 统计结束时间
}

func (m *StatisticsMessage) MsgID() uint16 { return MsgIDStatisticsMsg }

func (m *StatisticsMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 0, 60)
	buf = append(buf, m.VehicleColor)
	plate := []byte(m.VehiclePlate)
	for len(plate) < 21 {
		plate = append(plate, 0)
	}
	buf = append(buf, plate[:21]...)

	buf2 := make([]byte, 4)
	binary.BigEndian.PutUint32(buf2, m.TotalMsgCount)
	buf = append(buf, buf2...)

	binary.BigEndian.PutUint32(buf2, m.LocationCount)
	buf = append(buf, buf2...)

	binary.BigEndian.PutUint32(buf2, m.AlarmCount)
	buf = append(buf, buf2...)

	binary.BigEndian.PutUint32(buf2, m.OnlineDuration)
	buf = append(buf, buf2...)

	return buf, nil
}

func (m *StatisticsMessage) Unmarshal(data []byte) error {
	if len(data) < 38 {
		return fmt.Errorf("809 statistics too short")
	}
	m.VehicleColor = data[0]
	m.VehiclePlate = trimNull809(data[1:22])
	m.TotalMsgCount = binary.BigEndian.Uint32(data[22:26])
	m.LocationCount = binary.BigEndian.Uint32(data[26:30])
	m.AlarmCount = binary.BigEndian.Uint32(data[30:34])
	m.OnlineDuration = binary.BigEndian.Uint32(data[34:38])
	return nil
}

// ==================== 辅助函数 ====================

func encodeVehicleLocBody(color byte, plate, timeStr string, lat, lon float64, speed, direction uint16) []byte {
	buf := make([]byte, 0, 40)
	buf = append(buf, color)
	p := []byte(plate)
	for len(p) < 21 {
		p = append(p, 0)
	}
	buf = append(buf, p[:21]...)

	// 0报警 4字节 (填充0)
	buf = append(buf, 0, 0, 0, 0)

	latBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(latBytes, uint32(math.Abs(lat)*1000000))
	buf = append(buf, latBytes...)

	lonBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lonBytes, uint32(math.Abs(lon)*1000000))
	buf = append(buf, lonBytes...)

	spBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(spBytes, speed)
	buf = append(buf, spBytes...)

	dirBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(dirBytes, direction)
	buf = append(buf, dirBytes...)

	for len(timeStr) < 12 {
		timeStr = "0" + timeStr
	}
	for i := 0; i < 6; i++ {
		buf = append(buf, ((timeStr[i*2]-'0')<<4)|(timeStr[i*2+1]-'0'))
	}
	return buf
}

func decodeVehicleLocBody(data []byte) (byte, string, string, float64, float64, uint16, uint16, error) {
	if len(data) < 38 {
		return 0, "", "", 0, 0, 0, 0, fmt.Errorf("809 vehicle loc body too short: %d", len(data))
	}
	color := data[0]
	plate := trimNull809(data[1:22])
	lat := float64(binary.BigEndian.Uint32(data[26:30])) / 1000000.0
	lon := float64(binary.BigEndian.Uint32(data[30:34])) / 1000000.0
	speed := binary.BigEndian.Uint16(data[34:36])
	direction := binary.BigEndian.Uint16(data[36:38])
	timeStr := ""
	if len(data) >= 44 {
		t := make([]byte, 0, 12)
		for i := 0; i < 6; i++ {
			b := data[38+i]
			t = append(t, (b>>4)+'0', (b&0x0F)+'0')
		}
		timeStr = string(t)
	}
	return color, plate, timeStr, lat, lon, speed, direction, nil
}

func trimNull809(data []byte) string {
	end := len(data)
	for end > 0 && data[end-1] == 0 {
		end--
	}
	return string(data[:end])
}
