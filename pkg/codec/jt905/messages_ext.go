package jt905

import (
	"encoding/binary"
	"fmt"
	"math"
)

// ==================== JT/T 905-2014 完整消息ID ====================

const (
	MsgIDTaxiStatus      uint16 = 0x0B00 // 计价器状态
	MsgIDTaxiOperate     uint16 = 0x0B01 // 运营数据上报
	MsgIDTaxiDispatch    uint16 = 0x8B02 // 调度信息
	MsgIDTaxiPrice       uint16 = 0x8B03 // 计价信息
	MsgIDTaxiAd          uint16 = 0x8B04 // 广告信息
	// 扩展消息
	MsgIDTaxiLogin       uint16 = 0x0B05 // 出租车签到
	MsgIDTaxiLoginResp   uint16 = 0x8B05 // 签到应答
	MsgIDTaxiLogout      uint16 = 0x0B06 // 出租车签退
	MsgIDTaxiLogoutResp  uint16 = 0x8B06 // 签退应答
	MsgIDTaxiShift       uint16 = 0x0B07 // 交接班
	MsgIDTaxiShiftResp   uint16 = 0x8B07 // 交接班应答
	MsgIDTaxiEval        uint16 = 0x0B08 // 服务评价
	MsgIDTaxiEvalResp    uint16 = 0x8B08 // 服务评价应答
	MsgIDTaxiAlarm       uint16 = 0x0B09 // 出租车报警
	MsgIDTaxiAlarmResp   uint16 = 0x8B09 // 报警应答
	MsgIDTaxiPhone       uint16 = 0x8B0A // 电话回拨
	MsgIDTaxiSMS         uint16 = 0x8B0B // 短信下发
	MsgIDTaxiSMSResp     uint16 = 0x0B0B // 短信应答
	MsgIDTaxiInfoPush    uint16 = 0x8B0C // 信息推送
	MsgIDTaxiNavInfo     uint16 = 0x8B0D // 导航信息
	MsgIDTaxiOrder       uint16 = 0x0B0C // 网约车订单
	MsgIDTaxiOrderResp   uint16 = 0x8B0E // 订单下发
	MsgIDTaxiPay         uint16 = 0x0B0D // 支付上报
	MsgIDTaxiPayResp     uint16 = 0x8B0F // 支付应答
)

// MsgInfo905 905消息信息
type MsgInfo905 struct {
	ID   uint16
	Name string
	Dir  string
}

// AllMessages905 905协议全部消息列表
var AllMessages905 = []MsgInfo905{
	{0x0B00, "计价器状态", "up"},
	{0x0B01, "运营数据上报", "up"},
	{0x8B02, "调度信息", "down"},
	{0x8B03, "计价信息", "down"},
	{0x8B04, "广告信息", "down"},
	{0x0B05, "出租车签到", "up"},
	{0x8B05, "签到应答", "down"},
	{0x0B06, "出租车签退", "up"},
	{0x8B06, "签退应答", "down"},
	{0x0B07, "交接班", "up"},
	{0x8B07, "交接班应答", "down"},
	{0x0B08, "服务评价", "up"},
	{0x8B08, "服务评价应答", "down"},
	{0x0B09, "出租车报警", "up"},
	{0x8B09, "报警应答", "down"},
	{0x8B0A, "电话回拨", "down"},
	{0x8B0B, "短信下发", "down"},
	{0x0B0B, "短信应答", "up"},
	{0x8B0C, "信息推送", "down"},
	{0x8B0D, "导航信息", "down"},
	{0x0B0C, "网约车订单", "up"},
	{0x8B0E, "订单下发", "down"},
	{0x0B0D, "支付上报", "up"},
	{0x8B0F, "支付应答", "down"},
}

// TaxiStatusMessage 0x0B00 计价器状态
type TaxiStatusMessage struct {
	VehicleColor byte
	VehiclePlate string
	Status       byte   // 0=空车 1=载客 2=电招 3=暂停
	ChangeTime   string // 状态变化时间
	Lat          float64
	Lon          float64
}

func (m *TaxiStatusMessage) MsgID() uint16 { return MsgIDTaxiStatus }

func (m *TaxiStatusMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 0, 40)
	buf = append(buf, m.VehicleColor)
	plate := []byte(m.VehiclePlate)
	for len(plate) < 21 {
		plate = append(plate, 0)
	}
	buf = append(buf, plate[:21]...)
	buf = append(buf, m.Status)
	for len(m.ChangeTime) < 12 {
		m.ChangeTime = "0" + m.ChangeTime
	}
	for i := 0; i < 6; i++ {
		buf = append(buf, ((m.ChangeTime[i*2]-'0')<<4)|(m.ChangeTime[i*2+1]-'0'))
	}
	latBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(latBytes, uint32(math.Abs(m.Lat)*1000000))
	buf = append(buf, latBytes...)
	lonBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lonBytes, uint32(math.Abs(m.Lon)*1000000))
	buf = append(buf, lonBytes...)
	return buf, nil
}

func (m *TaxiStatusMessage) Unmarshal(data []byte) error {
	if len(data) < 37 {
		return fmt.Errorf("905 status too short")
	}
	m.VehicleColor = data[0]
	plate := data[1:22]
	for len(plate) > 0 && plate[len(plate)-1] == 0 {
		plate = plate[:len(plate)-1]
	}
	m.VehiclePlate = string(plate)
	m.Status = data[22]
	t := make([]byte, 0, 12)
	for i := 0; i < 6; i++ {
		b := data[23+i]
		t = append(t, (b>>4)+'0', (b&0x0F)+'0')
	}
	m.ChangeTime = string(t)
	m.Lat = float64(binary.BigEndian.Uint32(data[29:33])) / 1000000.0
	m.Lon = float64(binary.BigEndian.Uint32(data[33:37])) / 1000000.0
	return nil
}

// TaxiDispatchMessage 0x8B02 调度信息
type TaxiDispatchMessage struct {
	DispatchType byte   // 0=普通调度 1=电招
	Content      string // 调度内容
	PassengerPhone string // 乘客电话
	PickupAddr    string // 上车地点
}

func (m *TaxiDispatchMessage) MsgID() uint16 { return MsgIDTaxiDispatch }

func (m *TaxiDispatchMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 0, 100)
	buf = append(buf, m.DispatchType)
	content := []byte(m.Content)
	buf = append(buf, byte(len(content)))
	buf = append(buf, content...)
	phone := []byte(m.PassengerPhone)
	buf = append(buf, byte(len(phone)))
	buf = append(buf, phone...)
	addr := []byte(m.PickupAddr)
	buf = append(buf, byte(len(addr)))
	buf = append(buf, addr...)
	return buf, nil
}

func (m *TaxiDispatchMessage) Unmarshal(data []byte) error {
	if len(data) < 1 {
		return ErrDataTooShort905
	}
	m.DispatchType = data[0]
	offset := 1
	if offset >= len(data) { return nil }
	cl := int(data[offset]); offset++
	if offset+cl > len(data) { return nil }
	m.Content = string(data[offset:offset+cl]); offset += cl
	if offset >= len(data) { return nil }
	pl := int(data[offset]); offset++
	if offset+pl > len(data) { return nil }
	m.PassengerPhone = string(data[offset:offset+pl]); offset += pl
	if offset >= len(data) { return nil }
	al := int(data[offset]); offset++
	if offset+al > len(data) { return nil }
	m.PickupAddr = string(data[offset:offset+al])
	return nil
}

// TaxiPriceMessage 0x8B03 计价信息
type TaxiPriceMessage struct {
	StartTime    string // 计价开始时间
	UnitPrice    uint16 // 起步价(分)
	UnitDistance  uint16 // 起步距离(百米)
	PerKmPrice   uint16 // 每公里单价(分)
	NightSurcharge uint16 // 夜间附加费(分)
}

func (m *TaxiPriceMessage) MsgID() uint16 { return MsgIDTaxiPrice }

func (m *TaxiPriceMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 0, 20)
	for len(m.StartTime) < 12 {
		m.StartTime = "0" + m.StartTime
	}
	for i := 0; i < 6; i++ {
		buf = append(buf, ((m.StartTime[i*2]-'0')<<4)|(m.StartTime[i*2+1]-'0'))
	}
	priceBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(priceBytes, m.UnitPrice)
	buf = append(buf, priceBytes...)
	binary.BigEndian.PutUint16(priceBytes, m.UnitDistance)
	buf = append(buf, priceBytes...)
	binary.BigEndian.PutUint16(priceBytes, m.PerKmPrice)
	buf = append(buf, priceBytes...)
	binary.BigEndian.PutUint16(priceBytes, m.NightSurcharge)
	buf = append(buf, priceBytes...)
	return buf, nil
}

func (m *TaxiPriceMessage) Unmarshal(data []byte) error {
	if len(data) < 14 {
		return fmt.Errorf("905 price too short")
	}
	t := make([]byte, 0, 12)
	for i := 0; i < 6; i++ {
		b := data[i]
		t = append(t, (b>>4)+'0', (b&0x0F)+'0')
	}
	m.StartTime = string(t)
	m.UnitPrice = binary.BigEndian.Uint16(data[6:8])
	m.UnitDistance = binary.BigEndian.Uint16(data[8:10])
	m.PerKmPrice = binary.BigEndian.Uint16(data[10:12])
	m.NightSurcharge = binary.BigEndian.Uint16(data[12:14])
	return nil
}

// TaxiAdMessage 0x8B04 广告信息
type TaxiAdMessage struct {
	AdType    byte   // 广告类型
	PlayTimes byte   // 播放次数
	Content   string // 广告内容
}

func (m *TaxiAdMessage) MsgID() uint16 { return MsgIDTaxiAd }

func (m *TaxiAdMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 0, 200)
	buf = append(buf, m.AdType)
	buf = append(buf, m.PlayTimes)
	content := []byte(m.Content)
	buf = append(buf, byte(len(content)>>8), byte(len(content)&0xFF))
	buf = append(buf, content...)
	return buf, nil
}

func (m *TaxiAdMessage) Unmarshal(data []byte) error {
	if len(data) < 4 {
		return fmt.Errorf("905 ad too short")
	}
	m.AdType = data[0]
	m.PlayTimes = data[1]
	contentLen := int(binary.BigEndian.Uint16(data[2:4]))
	if 4+contentLen > len(data) {
		return fmt.Errorf("905 ad content too short")
	}
	m.Content = string(data[4 : 4+contentLen])
	return nil
}

var ErrDataTooShort905 = fmt.Errorf("905 data too short")
