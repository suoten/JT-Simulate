package gbt32960

import (
	"encoding/binary"
	"fmt"
)

// ==================== GB/T 32960.3-2016 完整消息类型 ====================

// MsgInfo32960 32960消息信息
type MsgInfo32960 struct {
	CmdFlag byte
	Name    string
}

// AllMessages32960 32960协议全部消息列表
var AllMessages32960 = []MsgInfo32960{
	{0x01, "车辆登入"},
	{0x02, "实时信息上报"},
	{0x03, "车辆登出"},
	{0x04, "平台登入"},
	{0x05, "平台登出"},
	{0x06, "校验应答"},
	{0x07, "心跳"},
}

// ==================== 实时信息数据项类型 ====================

const (
	DataItemVehicle      byte = 0x01 // 整车数据
	DataItemMotor        byte = 0x02 // 驱动电机数据
	DataItemFuelCell     byte = 0x03 // 燃料电池数据
	DataItemEngine       byte = 0x04 // 发动机数据
	DataItemPosition     byte = 0x05 // 位置数据
	DataItemExtreme      byte = 0x06 // 极值数据
	DataItemAlarm        byte = 0x07 // 报警数据
	DataItemBattery      byte = 0x08 // BMS电池数据
	DataItemHVSystem     byte = 0x09 // 高压系统数据
	DataItemBMSInfo      byte = 0x0A // BMS版本信息
	DataItemIBattery     byte = 0x0B // 蓄电池数据
	DataItemThermal      byte = 0x0C // 热管理数据
)

// VehicleData 整车数据
type VehicleData struct {
	Status       byte   // 车辆状态
	ChargeStatus byte   // 充电状态
	Mode         byte   // 模式
	Speed        uint16 // 车速(0.1km/h)
	TotalMileage uint32 // 累计里程(0.1km)
	Voltage      uint16 // 总电压(0.1V)
	Current      uint16 // 总电流(0.1A)
	SOC          byte   // SOC(%)
	DCDCStatus   byte   // DC-DC状态
	Shift        byte   // 档位
	Resistance   uint16 // 绝缘电阻(kΩ)
}

// EncodeVehicleData 编码整车数据
func EncodeVehicleData(d *VehicleData) []byte {
	buf := make([]byte, 0, 16)
	buf = append(buf, d.Status)
	buf = append(buf, d.ChargeStatus)
	buf = append(buf, d.Mode)
	spBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(spBytes, d.Speed)
	buf = append(buf, spBytes...)
	miBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(miBytes, d.TotalMileage)
	buf = append(buf, miBytes...)
	vBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(vBytes, d.Voltage)
	buf = append(buf, vBytes...)
	cBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(cBytes, d.Current)
	buf = append(buf, cBytes...)
	buf = append(buf, d.SOC)
	buf = append(buf, d.DCDCStatus)
	buf = append(buf, d.Shift)
	rBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(rBytes, d.Resistance)
	buf = append(buf, rBytes...)
	return buf
}

// PositionData 位置数据
type PositionData struct {
	ChargeStatus byte   // 充电状态
	Lat          int32  // 纬度(度*10^6)
	Lon          int32  // 经度(度*10^6)
	Altitude     uint16 // 海拔(0.1m)
	Direction    uint16 // 方向(0-359)
	Speed        uint16 // 速度(0.1km/h)
}

// EncodePositionData 编码位置数据
func EncodePositionData(d *PositionData) []byte {
	buf := make([]byte, 0, 14)
	buf = append(buf, d.ChargeStatus)
	latBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(latBytes, uint32(d.Lat))
	buf = append(buf, latBytes...)
	lonBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lonBytes, uint32(d.Lon))
	buf = append(buf, lonBytes...)
	altBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(altBytes, d.Altitude)
	buf = append(buf, altBytes...)
	dirBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(dirBytes, d.Direction)
	buf = append(buf, dirBytes...)
	spBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(spBytes, d.Speed)
	buf = append(buf, spBytes...)
	return buf
}

// AlarmData 报警数据
type AlarmData struct {
	MaxAlarmLevel byte   // 最高报警等级
	AlarmFlag     uint32 // 通用报警标志
	SubsysCount   byte   // 故障总数
	Alarms        []SubsysAlarm
}

type SubsysAlarm struct {
	SubsysID  byte   // 子系统编号
	AlarmCode byte   // 故障代码
	AlarmLevel byte   // 故障等级
}

// EncodeAlarmData 编码报警数据
func EncodeAlarmData(d *AlarmData) []byte {
	buf := make([]byte, 0, 6+len(d.Alarms)*3)
	buf = append(buf, d.MaxAlarmLevel)
	flagBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(flagBytes, d.AlarmFlag)
	buf = append(buf, flagBytes...)
	buf = append(buf, byte(len(d.Alarms)))
	for _, a := range d.Alarms {
		buf = append(buf, a.SubsysID)
		buf = append(buf, a.AlarmCode)
		buf = append(buf, a.AlarmLevel)
	}
	return buf
}

// EncodeRealtimeInfo 编码实时信息上报
func EncodeRealtimeInfo(vin string, collectTime string, seqNum uint16, items []RealtimeDataItem) []byte {
	body := make([]byte, 0, 20)
	for len(collectTime) < 12 {
		collectTime = "0" + collectTime
	}
	for i := 0; i < 6; i++ {
		body = append(body, ((collectTime[i*2]-'0')<<4)|(collectTime[i*2+1]-'0'))
	}
	seq := make([]byte, 2)
	binary.BigEndian.PutUint16(seq, seqNum)
	body = append(body, seq...)

	itemCount := make([]byte, 2)
	binary.BigEndian.PutUint16(itemCount, uint16(len(items)))
	body = append(body, itemCount...)

	for _, item := range items {
		body = append(body, item.Type)
		body = append(body, byte(len(item.Data)>>8), byte(len(item.Data)&0xFF))
		body = append(body, item.Data...)
	}

	h := &Header{
		StartChar:   0x23,
		CmdFlag:     MsgTypeRealtimeInfo,
		RespFlag:    0xFE,
		VIN:         vin,
		EncryptType: 0x01,
	}
	return Encode(h, body)
}

// RealtimeDataItem 实时数据项
type RealtimeDataItem struct {
	Type byte
	Data []byte
}

// EncodePlatformLogin 编码平台登入
func EncodePlatformLogin(platformID string, loginTime string, password string, encryptKey string) []byte {
	body := make([]byte, 0, 40)
	// 平台唯一标识 11字节
	pid := []byte(platformID)
	for len(pid) < 11 {
		pid = append(pid, 0)
	}
	body = append(body, pid[:11]...)
	// 时间 6字节BCD
	for len(loginTime) < 12 {
		loginTime = "0" + loginTime
	}
	for i := 0; i < 6; i++ {
		body = append(body, ((loginTime[i*2]-'0')<<4)|(loginTime[i*2+1]-'0'))
	}
	// 密码 20字节
	pwd := []byte(password)
	for len(pwd) < 20 {
		pwd = append(pwd, 0)
	}
	body = append(body, pwd[:20]...)
	// 加密密钥 16字节
	key := []byte(encryptKey)
	for len(key) < 16 {
		key = append(key, 0)
	}
	body = append(body, key[:16]...)

	h := &Header{
		StartChar:   0x23,
		CmdFlag:     MsgTypePlatformLogin,
		RespFlag:    0xFE,
		VIN:         "                 ", // 平台登入VIN为空格
		EncryptType: 0x01,
	}
	return Encode(h, body)
}

// EncodeCheckResp 编码校验应答
func EncodeCheckResp(vin string, cmdFlag byte, respTime string, seqNum uint16, result byte) []byte {
	body := make([]byte, 0, 10)
	body = append(body, cmdFlag)
	for len(respTime) < 12 {
		respTime = "0" + respTime
	}
	for i := 0; i < 6; i++ {
		body = append(body, ((respTime[i*2]-'0')<<4)|(respTime[i*2+1]-'0'))
	}
	seq := make([]byte, 2)
	binary.BigEndian.PutUint16(seq, seqNum)
	body = append(body, seq...)
	body = append(body, result)

	h := &Header{
		StartChar:   0x23,
		CmdFlag:     MsgTypeCheckResp,
		RespFlag:    0x01,
		VIN:         vin,
		EncryptType: 0x01,
	}
	return Encode(h, body)
}

// EncodeHeartbeat 编码心跳
func EncodeHeartbeat(vin string) []byte {
	h := &Header{
		StartChar:   0x23,
		CmdFlag:     MsgTypeHeartbeat,
		RespFlag:    0xFE,
		VIN:         vin,
		EncryptType: 0x01,
	}
	return Encode(h, nil)
}

// EncodeVehicleLogout 编码车辆登出
func EncodeVehicleLogout(vin string, logoutTime string, seqNum uint16) []byte {
	body := make([]byte, 0, 10)
	for len(logoutTime) < 12 {
		logoutTime = "0" + logoutTime
	}
	for i := 0; i < 6; i++ {
		body = append(body, ((logoutTime[i*2]-'0')<<4)|(logoutTime[i*2+1]-'0'))
	}
	seq := make([]byte, 2)
	binary.BigEndian.PutUint16(seq, seqNum)
	body = append(body, seq...)

	h := &Header{
		StartChar:   0x23,
		CmdFlag:     MsgTypeVehicleLogout,
		RespFlag:    0xFE,
		VIN:         vin,
		EncryptType: 0x01,
	}
	return Encode(h, body)
}

// MsgTypeStringAll 返回所有32960消息类型名称（扩展版）
func MsgTypeStringAll(cmdFlag byte) string {
	for _, m := range AllMessages32960 {
		if m.CmdFlag == cmdFlag {
			return m.Name
		}
	}
	return fmt.Sprintf("未知类型(0x%02X)", cmdFlag)
}
