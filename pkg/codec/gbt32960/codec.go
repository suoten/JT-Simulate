package gbt32960

import (
	"encoding/binary"
	"fmt"
)

// GB/T 32960.3-2016 新能源汽车协议消息类型（命令标识）
const (
	MsgTypeVehicleLogin    byte = 0x01 // 车辆登入
	MsgTypeRealtimeInfo    byte = 0x02 // 实时信息上报
	MsgTypeVehicleLogout   byte = 0x03 // 车辆登出
	MsgTypePlatformLogin   byte = 0x04 // 平台登入
	MsgTypePlatformLogout  byte = 0x05 // 平台登出
	MsgTypeCheckResp       byte = 0x06 // 校验应答（平台对终端的应答）
	MsgTypeHeartbeat       byte = 0x07 // 心跳
)

// 应答标志
const (
	RespSuccess byte = 0x01
	RespError   byte = 0x02
)

// Header 32960 报文头
type Header struct {
	StartChar  byte   // 起始符 0x23
	CmdFlag    byte   // 命令标识
	RespFlag   byte   // 应答标志
	VIN        string // 车辆识别码 (17字节)
	EncryptType byte  // 加密方式
	BodyLen    uint16 // 数据长度
}

// Packet 32960 完整报文
type Packet struct {
	Header Header
	Body   []byte
	BCC    byte // 校验码
}

// Encode 编码报文
func Encode(h *Header, body []byte) []byte {
	h.BodyLen = uint16(len(body))

	buf := make([]byte, 0, 24+len(body)+1)
	buf = append(buf, h.StartChar)
	buf = append(buf, h.CmdFlag)
	buf = append(buf, h.RespFlag)
	vin := []byte(h.VIN)
	for len(vin) < 17 {
		vin = append(vin, 0)
	}
	buf = append(buf, vin[:17]...)
	buf = append(buf, h.EncryptType)

	lenBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(lenBytes, h.BodyLen)
	buf = append(buf, lenBytes...)

	buf = append(buf, body...)

	// BCC校验
	var bcc byte
	for _, b := range buf {
		bcc ^= b
	}
	buf = append(buf, bcc)

	return buf
}

// Decode 解码报文
func Decode(data []byte) (*Packet, error) {
	if len(data) < 24 {
		return nil, fmt.Errorf("32960 packet too short: %d", len(data))
	}
	if data[0] != 0x23 {
		return nil, fmt.Errorf("invalid start char: 0x%02X", data[0])
	}

	pkt := &Packet{}
	pkt.Header.StartChar = data[0]
	pkt.Header.CmdFlag = data[1]
	pkt.Header.RespFlag = data[2]
	pkt.Header.VIN = string(data[3:20])
	pkt.Header.EncryptType = data[20]
	pkt.Header.BodyLen = binary.BigEndian.Uint16(data[21:23])

	if len(data) < int(23+pkt.Header.BodyLen)+1 {
		return nil, fmt.Errorf("32960 packet body too short")
	}

	pkt.Body = make([]byte, pkt.Header.BodyLen)
	copy(pkt.Body, data[23:23+pkt.Header.BodyLen])
	pkt.BCC = data[23+pkt.Header.BodyLen]

	// BCC校验
	var bcc byte
	for _, b := range data[:23+pkt.Header.BodyLen] {
		bcc ^= b
	}
	if bcc != pkt.BCC {
		return nil, fmt.Errorf("32960 BCC checksum error")
	}

	return pkt, nil
}

// VehicleLoginBody 车辆登入报文体
type VehicleLoginBody struct {
	LoginTime   string // YYMMDDHHmmss
	SeqNum      uint16
	ICCID       string // SIM卡ICCID (20字节)
	Subsystem   byte   // 充电子系统数
	SystemLen   byte
	SoftwareVer string
}

// EncodeVehicleLogin 编码车辆登入
func EncodeVehicleLogin(vin string, loginTime string, iccid string, seqNum uint16) []byte {
	body := make([]byte, 0, 30)
	// 时间BCD 6字节
	t := loginTime
	for len(t) < 12 {
		t = "0" + t
	}
	for i := 0; i < 6; i++ {
		high := t[i*2] - '0'
		low := t[i*2+1] - '0'
		body = append(body, (high<<4)|low)
	}
	seq := make([]byte, 2)
	binary.BigEndian.PutUint16(seq, seqNum)
	body = append(body, seq...)

	ic := []byte(iccid)
	for len(ic) < 20 {
		ic = append(ic, 0)
	}
	body = append(body, ic[:20]...)
	body = append(body, 1)    // 充电子系统数=1
	body = append(body, 1)    // 系统编码长度=1
	body = append(body, 0x01) // 系统编码
	// 软件版本
	ver := []byte("V1.0.0")
	body = append(body, byte(len(ver)))
	body = append(body, ver...)

	h := &Header{
		StartChar:   0x23,
		CmdFlag:     MsgTypeVehicleLogin,
		RespFlag:    0xFE, // 无应答
		VIN:         vin,
		EncryptType: 0x01, // 不加密
	}
	return Encode(h, body)
}

// RealtimeInfoBody 实时信息
type RealtimeInfoBody struct {
	CollectTime  string
	SeqNum       uint16
	TotalItems   uint16
	Items        []DataItem
}

type DataItem struct {
	Type   byte
	Value  []byte
}

// MsgTypeString 消息类型名称
func MsgTypeString(cmdFlag byte) string {
	return MsgTypeStringAll(cmdFlag)
}
