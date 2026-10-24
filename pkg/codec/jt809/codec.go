package jt809

import (
	"encoding/binary"
	"fmt"
	"time"
)

// 帧首尾标记（与 SmallChi/JT809 实现一致）
const (
	BeginFlag = 0x5B
	EndFlag   = 0x5D
)

// 帧类型
const (
	FrameTypeUp   = 0 // 下级→上级
	FrameTypeDown = 1 // 上级→下级
)

// 消息头长度
// JT/T 809-2011: 22 字节
// MsgLength(4) + MsgSN(4) + BusinessType(2) + MsgGNSSCENTERID(4) + Version(3) + EncryptFlag(1) + EncryptKey(4)
// JT/T 809-2019: 30 字节（在 2011 基础上增加 UTCDateTime(8)）
const (
	HeaderLen2011 = 22
	HeaderLen2019 = 30
	HeaderLen     = HeaderLen2019 // 默认使用 2019 版头长度
)

// 校验码长度（CRC16，2 字节）
const checksumLen = 2

// 消息头
type Header struct {
	MsgLength    uint32 // 消息总长度（含头、体、校验码、终止符）
	MsgSN        uint32 // 消息流水号
	MsgID        uint16 // 消息ID（业务类型标识）
	MsgGNSSCENTERID uint32 // 下级平台接入码/服务中心ID
	Version      []byte // 协议版本号标识（3 字节）
	MsgEncrypt   uint8  // 报文加密标识
	EncryptKey   uint32 // 数据加密密钥
	UTCTime      uint64 // 2019 版新增：发送消息时的系统 UTC 时间（8 字节）
	VersionFlag  uint8  // 版本标识（内部使用：0=2011, 1=2019）
	BodyLength   uint16 // 消息体长度（从 MsgLength 计算）
	VehicleColor byte   // 车牌颜色（仅业务消息）
	VehiclePlate string // 车牌号码（仅业务消息，21字节）
}

// JT809Codec JT809编解码器
type JT809Codec struct{}

func NewCodec() *JT809Codec {
	return &JT809Codec{}
}

// headerLenByVersion 根据版本标识返回消息头长度
func headerLenByVersion(versionFlag uint8) int {
	if versionFlag == 0 {
		return HeaderLen2011 // 2011 版，22 字节
	}
	return HeaderLen2019 // 2019 版，30 字节
}

// plausibleUTCTimestamp 判断 8 字节大端整数是否为合理的 Unix 时间戳。
// 本实现 Encode 写入的是毫秒级时间戳（UnixMilli），标准对端可能使用秒级，
// 因此两种量级均接受；范围放宽到 2015-01-01 ~ 2100-01-01。
func plausibleUTCTimestamp(b []byte) bool {
	if len(b) < 8 {
		return false
	}
	v := binary.BigEndian.Uint64(b)
	const (
		minSeconds uint64 = 1420070400    // 2015-01-01
		maxSeconds uint64 = 4102444800    // 2100-01-01
		minMillis  uint64 = 1420070400000 // 2015-01-01 (ms)
		maxMillis  uint64 = 4102444800000 // 2100-01-01 (ms)
	)
	return (v >= minSeconds && v <= maxSeconds) || (v >= minMillis && v <= maxMillis)
}

// ParseHeader 解析消息头（自动检测 2011/2019 版本）
func (c *JT809Codec) ParseHeader(data []byte) (*Header, int, error) {
	if len(data) < HeaderLen2011 {
		return nil, 0, fmt.Errorf("809 header too short: %d bytes (minimum %d)", len(data), HeaderLen2011)
	}
	h := &Header{}
	h.MsgLength = binary.BigEndian.Uint32(data[0:4])
	h.MsgSN = binary.BigEndian.Uint32(data[4:8])
	h.MsgID = binary.BigEndian.Uint16(data[8:10])
	h.MsgGNSSCENTERID = binary.BigEndian.Uint32(data[10:14])
	h.Version = make([]byte, 3)
	copy(h.Version, data[14:17])
	h.MsgEncrypt = data[17]
	h.EncryptKey = binary.BigEndian.Uint32(data[18:22])

	// 版本检测策略：
	// 1. 优先通过 Version 字段第一个字节判断：0x01 = 2019 版
	// 2. 如果 Version[0] != 0x01，按长度无法区分 2011/2019：
	//    两种头长假设都与 MsgLength 自洽（2011 版只是把 2019 版头部的
	//    8 字节 UTC 时间戳当作消息体前缀），因此借助 2019 版头部特有的
	//    UTC 时间戳字段做兜底：MsgLength 与数据长度吻合且 data[22:30]
	//    是合理的 Unix 时间戳（秒/毫秒）时判定为 2019 版。
	// 3. 否则回退到 2011 版
	is2019 := false
	if len(h.Version) >= 1 && h.Version[0] == 0x01 {
		is2019 = true
	} else if len(data) >= HeaderLen2019 {
		bodyLen2019 := int(h.MsgLength) - HeaderLen2019 - checksumLen - 1 // -1 for endflag
		if bodyLen2019 >= 0 && HeaderLen2019+bodyLen2019 == len(data) && plausibleUTCTimestamp(data[22:30]) {
			is2019 = true
		}
	}

	if is2019 && len(data) >= HeaderLen2019 {
		h.VersionFlag = 1
		h.UTCTime = binary.BigEndian.Uint64(data[22:30])
		bodyLen2019 := int(h.MsgLength) - HeaderLen2019 - checksumLen - 1
		if bodyLen2019 < 0 {
			bodyLen2019 = int(h.MsgLength) - HeaderLen2019 - checksumLen
		}
		h.BodyLength = uint16(bodyLen2019)
		return h, HeaderLen2019, nil
	}

	// 2011 版：头长 22 字节
	h.VersionFlag = 0
	bodyLen2011 := int(h.MsgLength) - HeaderLen2011 - checksumLen - 1 // -1 for endflag
	if bodyLen2011 < 0 {
		// 可能 MsgLength 不含 endflag
		bodyLen2011 = int(h.MsgLength) - HeaderLen2011 - checksumLen
		if bodyLen2011 < 0 {
			return nil, 0, fmt.Errorf("809 invalid MsgLength %d for header %d", h.MsgLength, HeaderLen2011)
		}
	}
	h.BodyLength = uint16(bodyLen2011)
	return h, HeaderLen2011, nil
}

// EncodeHeader 编码消息头
func (c *JT809Codec) EncodeHeader(h *Header) ([]byte, error) {
	headerLen := headerLenByVersion(h.VersionFlag)
	buf := make([]byte, headerLen)
	binary.BigEndian.PutUint32(buf[0:4], h.MsgLength)
	binary.BigEndian.PutUint32(buf[4:8], h.MsgSN)
	binary.BigEndian.PutUint16(buf[8:10], h.MsgID)
	binary.BigEndian.PutUint32(buf[10:14], h.MsgGNSSCENTERID)

	// 版本号标识（3 字节）：保留调用方已提供的字节，后位补零；
	// 仅在未提供任何版本字节且为 2019 版时使用默认版本号 0x01
	ver := make([]byte, 3)
	copy(ver, h.Version)
	if len(h.Version) == 0 && h.VersionFlag == 1 {
		ver[0] = 0x01 // 2019 版版本号
	}
	copy(buf[14:17], ver)

	buf[17] = h.MsgEncrypt
	binary.BigEndian.PutUint32(buf[18:22], h.EncryptKey)

	if headerLen == HeaderLen2019 {
		// 2019 版：UTC 时间戳（8 字节）
		binary.BigEndian.PutUint64(buf[22:30], h.UTCTime)
	}
	return buf, nil
}

// calcCRC16 计算 CRC16 校验码（与 SmallChi/JT809 的 WriteCRC16 一致）
// 使用 CRC-16/XMODEM 算法
func calcCRC16(data []byte) uint16 {
	crc := uint16(0)
	for _, b := range data {
		crc ^= uint16(b) << 8
		for i := 0; i < 8; i++ {
			if crc&0x8000 != 0 {
				crc = (crc << 1) ^ 0x1021
			} else {
				crc <<= 1
			}
		}
	}
	return crc
}

// Encode 编码完整消息
func (c *JT809Codec) Encode(h *Header, body []byte) ([]byte, error) {
	headerLen := headerLenByVersion(h.VersionFlag)
	bodyLen := len(body)
	h.BodyLength = uint16(bodyLen)
	// MsgLength = header + body + CRC16(2) + endflag(1)
	h.MsgLength = uint32(headerLen + bodyLen + checksumLen + 1)

	// 2019 版设置 UTC 时间戳
	if h.VersionFlag == 1 && h.UTCTime == 0 {
		h.UTCTime = uint64(time.Now().UnixMilli())
	}

	headerBytes, err := c.EncodeHeader(h)
	if err != nil {
		return nil, err
	}

	// 组装：头 + 体
	raw := make([]byte, 0, headerLen+bodyLen+checksumLen)
	raw = append(raw, headerBytes...)
	raw = append(raw, body...)

	// 校验码（CRC16，2 字节，大端）
	crc := calcCRC16(raw)
	raw = append(raw, byte(crc>>8), byte(crc&0xFF))

	// 转义（JT/T 809 标准，与 SmallChi/JT809 一致）：
	// 0x5B → 0x5A 0x01；0x5A → 0x5A 0x02；0x5D → 0x5E 0x01；0x5E → 0x5E 0x02
	// 注意：帧首 0x5B 和帧尾 0x5D 是标记，不在转义范围内；
	// CRC 已在上面对未转义数据计算完成
	escaped := make([]byte, 0, len(raw)*2)
	for _, b := range raw {
		switch b {
		case 0x5B:
			escaped = append(escaped, 0x5A, 0x01)
		case 0x5A:
			escaped = append(escaped, 0x5A, 0x02)
		case 0x5D:
			escaped = append(escaped, 0x5E, 0x01)
		case 0x5E:
			escaped = append(escaped, 0x5E, 0x02)
		default:
			escaped = append(escaped, b)
		}
	}

	// 添加首尾标记
	result := make([]byte, 0, len(escaped)+2)
	result = append(result, BeginFlag)  // 0x5B
	result = append(result, escaped...)
	result = append(result, EndFlag)    // 0x5D

	return result, nil
}

// Decode 解码完整消息
func (c *JT809Codec) Decode(frame []byte) (*Header, []byte, error) {
	if len(frame) < 2 {
		return nil, nil, fmt.Errorf("frame too short")
	}
	if frame[0] != BeginFlag || frame[len(frame)-1] != EndFlag {
		return nil, nil, fmt.Errorf("invalid 809 delimiter: expected 0x5B...0x5D, got 0x%02X...0x%02X", frame[0], frame[len(frame)-1])
	}

	// 去除首尾标记
	striped := frame[1 : len(frame)-1]

	// 反转义（JT/T 809 标准，与 SmallChi/JT809 一致）：
	// 0x5A 0x01 → 0x5B；0x5A 0x02 → 0x5A；0x5E 0x01 → 0x5D；0x5E 0x02 → 0x5E
	raw := make([]byte, 0, len(striped))
	i := 0
	for i < len(striped) {
		if (striped[i] == 0x5A || striped[i] == 0x5E) && i+1 < len(striped) {
			switch {
			case striped[i] == 0x5A && striped[i+1] == 0x01:
				raw = append(raw, 0x5B)
			case striped[i] == 0x5A && striped[i+1] == 0x02:
				raw = append(raw, 0x5A)
			case striped[i] == 0x5E && striped[i+1] == 0x01:
				raw = append(raw, 0x5D)
			case striped[i] == 0x5E && striped[i+1] == 0x02:
				raw = append(raw, 0x5E)
			default:
				// 不是合法的转义序列，保留原始字节
				raw = append(raw, striped[i], striped[i+1])
			}
			i += 2
		} else {
			raw = append(raw, striped[i])
			i++
		}
	}

	// 校验码校验（CRC16，2 字节）
	if len(raw) < checksumLen {
		return nil, nil, fmt.Errorf("raw data too short for checksum")
	}
	dataWithoutCRC := raw[:len(raw)-checksumLen]
	crcInFrame := binary.BigEndian.Uint16(raw[len(raw)-checksumLen:])
	crcCalc := calcCRC16(dataWithoutCRC)
	if crcInFrame != crcCalc {
		return nil, nil, fmt.Errorf("invalid 809 CRC16: expected 0x%04X, got 0x%04X", crcCalc, crcInFrame)
	}

	// 去除校验码
	data := dataWithoutCRC

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
