package jt809

import (
	"encoding/binary"
	"fmt"
)

// 帧类型
const (
	FrameTypeUp   = 0 // 下级→上级
	FrameTypeDown = 1 // 上级→下级
)

// 消息头长度
// JT/T 809-2019: 消息头 = 消息长度(4) + 消息流水号(4) + 消息ID(2) + 消息体属性(2) + 消息体长度(2) + 终端手机号(6) + 消息流水号(2) ... 实际809头为22字节
const HeaderLen = 22

// 消息头
type Header struct {
	MsgLength    uint32 // 消息总长度（含头、体、校验码）
	MsgSN        uint32 // 消息流水号
	MsgID        uint16 // 消息ID
	MsgEncrypt   uint8  // 消息体属性（加密方式 bit0-2, bit3-7保留）
	EncryptKey   uint32 // 加密密钥ID
	GNSSCenterID uint32 // 服务中心ID
	VersionFlag  uint8  // 版本标识
	BodyLength   uint16 // 消息体长度（从BodyLength字段解析）
	VehicleColor byte   // 车牌颜色（仅业务消息）
	VehiclePlate string // 车牌号码（仅业务消息，21字节）
}

// JT809Codec JT809编解码器
type JT809Codec struct{}

func NewCodec() *JT809Codec {
	return &JT809Codec{}
}

// ParseHeader 解析消息头
func (c *JT809Codec) ParseHeader(data []byte) (*Header, int, error) {
	if len(data) < HeaderLen {
		return nil, 0, fmt.Errorf("809 header too short: %d bytes", len(data))
	}
	h := &Header{}
	h.MsgLength = binary.BigEndian.Uint32(data[0:4])
	h.MsgSN = binary.BigEndian.Uint32(data[4:8])
	h.MsgID = binary.BigEndian.Uint16(data[8:10])
	h.MsgEncrypt = data[10]
	// data[11] is reserved
	h.EncryptKey = binary.BigEndian.Uint32(data[12:16])
	h.GNSSCenterID = binary.BigEndian.Uint32(data[16:20])
	h.VersionFlag = data[20]
	// body length is derived
	h.BodyLength = uint16(h.MsgLength) - uint16(HeaderLen) - 1 // -1 for checksum
	return h, HeaderLen, nil
}

// EncodeHeader 编码消息头
func (c *JT809Codec) EncodeHeader(h *Header) ([]byte, error) {
	buf := make([]byte, HeaderLen)
	binary.BigEndian.PutUint32(buf[0:4], h.MsgLength)
	binary.BigEndian.PutUint32(buf[4:8], h.MsgSN)
	binary.BigEndian.PutUint16(buf[8:10], h.MsgID)
	buf[10] = h.MsgEncrypt
	// buf[11] reserved
	binary.BigEndian.PutUint32(buf[12:16], h.EncryptKey)
	binary.BigEndian.PutUint32(buf[16:20], h.GNSSCenterID)
	buf[20] = h.VersionFlag
	return buf, nil
}

// Encode 编码完整消息
func (c *JT809Codec) Encode(h *Header, body []byte) ([]byte, error) {
	bodyLen := len(body)
	h.BodyLength = uint16(bodyLen)
	h.MsgLength = uint32(HeaderLen + bodyLen + 1) // +1 for checksum

	headerBytes, err := c.EncodeHeader(h)
	if err != nil {
		return nil, err
	}

	// 组装：头 + 体
	raw := make([]byte, 0, int(h.MsgLength))
	raw = append(raw, headerBytes...)
	raw = append(raw, body...)

	// 校验码（异或）
	checksum := byte(0)
	for _, b := range raw {
		checksum ^= b
	}
	raw = append(raw, checksum)

	// 转义：0x5B → 0x5B 0x5E, 0x5E → 0x5B 0x5D, 0x5A → 0x5B 0x5C
	escaped := make([]byte, 0, len(raw)*2)
	for _, b := range raw {
		switch b {
		case 0x5B:
			escaped = append(escaped, 0x5B, 0x5E)
		case 0x5E:
			escaped = append(escaped, 0x5B, 0x5D)
		case 0x5A:
			escaped = append(escaped, 0x5B, 0x5C)
		default:
			escaped = append(escaped, b)
		}
	}

	// 添加首尾标记
	result := make([]byte, 0, len(escaped)+2)
	result = append(result, 0x5B)
	result = append(result, escaped...)
	result = append(result, 0x5E)

	return result, nil
}

// Decode 解码完整消息
func (c *JT809Codec) Decode(frame []byte) (*Header, []byte, error) {
	if len(frame) < 2 {
		return nil, nil, fmt.Errorf("frame too short")
	}
	if frame[0] != 0x5B || frame[len(frame)-1] != 0x5E {
		return nil, nil, fmt.Errorf("invalid 809 delimiter")
	}

	// 去除首尾标记
	striped := frame[1 : len(frame)-1]

	// 反转义
	raw := make([]byte, 0, len(striped))
	i := 0
	for i < len(striped) {
		if striped[i] == 0x5B && i+1 < len(striped) {
			switch striped[i+1] {
			case 0x5E:
				raw = append(raw, 0x5B)
			case 0x5D:
				raw = append(raw, 0x5E)
			case 0x5C:
				raw = append(raw, 0x5A)
			default:
				raw = append(raw, striped[i], striped[i+1])
			}
			i += 2
		} else {
			raw = append(raw, striped[i])
			i++
		}
	}

	// 校验码校验
	if len(raw) < 2 {
		return nil, nil, fmt.Errorf("raw data too short")
	}
	var xor byte
	for j := 0; j < len(raw)-1; j++ {
		xor ^= raw[j]
	}
	if xor != raw[len(raw)-1] {
		return nil, nil, fmt.Errorf("invalid 809 checksum")
	}

	// 去除校验码
	data := raw[:len(raw)-1]

	// 解析消息头
	h, headerLen, err := c.ParseHeader(data)
	if err != nil {
		return nil, nil, err
	}

	// 提取消息体
	bodyData := data[headerLen:]
	return h, bodyData, nil
}

// MsgName 消息名称
func MsgName(msgID uint16) string {
	for _, m := range AllMessages809 {
		if m.ID == msgID {
			return m.Name
		}
	}
	return fmt.Sprintf("未知消息(0x%04X)", msgID)
}
