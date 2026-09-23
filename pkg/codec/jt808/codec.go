package jt808

import (
	"encoding/binary"
	"fmt"

	"github.com/suoten/jt-simulate/pkg/types"
)

// JT808Codec JT808编解码器
type JT808Codec struct{}

// NewCodec 创建编解码器
func NewCodec() *JT808Codec {
	return &JT808Codec{}
}

func (c *JT808Codec) ProtocolType() types.ProtocolType {
	return types.ProtocolJT808
}

// ParseHeader 解析消息头
func (c *JT808Codec) ParseHeader(data []byte) (*types.MessageHeader, int, error) {
	if len(data) < 12 {
		return nil, 0, fmt.Errorf("header too short: %d bytes", len(data))
	}

	header := &types.MessageHeader{}
	header.MsgID = binary.BigEndian.Uint16(data[0:2])
	header.BodyAttr = binary.BigEndian.Uint16(data[2:4])

	header.BodyLen = int(header.BodyAttr & 0x03FF)
	header.EncryptMethod = uint8((header.BodyAttr >> 10) & 0x07)
	header.Version2019 = (header.BodyAttr & 0x8000) != 0
	header.HasPack = (header.BodyAttr & 0x2000) != 0

	phoneStart := 4
	if header.Version2019 {
		if len(data) < 13 {
			return nil, 0, fmt.Errorf("2019 header too short: %d bytes", len(data))
		}
		header.ProtocolVer = data[4]
		phoneStart = 5
	}

	phoneBCD := data[phoneStart : phoneStart+6]
	header.Phone = BCDToStringSafe(phoneBCD)
	header.SeqNum = binary.BigEndian.Uint16(data[phoneStart+6 : phoneStart+8])

	offset := phoneStart + 8
	if header.HasPack {
		if len(data) < offset+4 {
			return nil, 0, fmt.Errorf("header with pack info too short")
		}
		header.PackTotal = binary.BigEndian.Uint16(data[offset : offset+2])
		header.PackIndex = binary.BigEndian.Uint16(data[offset+2 : offset+4])
		offset += 4
	}

	return header, offset, nil
}

// EncodeHeader 编码消息头
func (c *JT808Codec) EncodeHeader(header *types.MessageHeader) ([]byte, error) {
	buf := make([]byte, 0, 21)

	msgIDBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(msgIDBytes, header.MsgID)
	buf = append(buf, msgIDBytes...)

	bodyAttr := uint16(header.BodyLen) & 0x03FF
	bodyAttr |= (uint16(header.EncryptMethod) & 0x07) << 10
	if header.Version2019 {
		bodyAttr |= 0x8000
	}
	if header.HasPack {
		bodyAttr |= 0x2000
	}
	bodyAttrBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(bodyAttrBytes, bodyAttr)
	buf = append(buf, bodyAttrBytes...)

	if header.Version2019 {
		buf = append(buf, header.ProtocolVer)
	}

	phoneBCD, err := StringToBCD6(header.Phone)
	if err != nil {
		return nil, fmt.Errorf("encode header: %w", err)
	}
	buf = append(buf, phoneBCD...)

	seqBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(seqBytes, header.SeqNum)
	buf = append(buf, seqBytes...)

	if header.HasPack {
		packTotal := make([]byte, 2)
		binary.BigEndian.PutUint16(packTotal, header.PackTotal)
		buf = append(buf, packTotal...)

		packIndex := make([]byte, 2)
		binary.BigEndian.PutUint16(packIndex, header.PackIndex)
		buf = append(buf, packIndex...)
	}

	return buf, nil
}

// ParseBody 解析消息体
func (c *JT808Codec) ParseBody(msgID uint16, data []byte) (types.MessageBody, error) {
	var body types.MessageBody

	switch msgID {
	case MsgIDTerminalGeneralResp:
		body = &TerminalGeneralRespMessage{}
	case 0x8001:
		body = &PlatformGeneralRespMessage{}
	case MsgIDHeartbeat:
		body = &HeartbeatMessage{}
	case MsgIDRegister:
		body = &RegisterMessage{}
	case MsgIDRegisterResp:
		body = &RegisterRespMessage{}
	case MsgIDAuth:
		body = &AuthMessage{}
	case MsgIDLocation:
		body = &LocationMessage{}
	case MsgIDLocationBatch:
		body = &LocationBatchMessage{}
	case MsgIDCommand:
		body = &CommandMessage{}
	case MsgIDCommandResp:
		body = &CommandRespMessage{}
	case MsgIDParamQuery:
		body = &ParamQueryMessage{}
	case MsgIDParamResp:
		body = &ParamRespMessage{}
	case MsgIDTerminalCtrl:
		body = &TerminalCtrlMessage{}
	case MsgIDLocationQuery:
		body = &LocationQueryMessage{}
	case MsgIDTempLocationTrack:
		body = &TempLocationTrackMessage{}
	case MsgIDManualAlarmConfirm:
		body = &ManualAlarmConfirmMessage{}
	case MsgIDTextSend:
		body = &TextSendMessage{}
	case MsgIDOverspeedSet:
		body = &OverspeedSetMessage{}
	case MsgIDFatigueDriveSet:
		body = &FatigueDriveSetMessage{}
	case MsgIDVehicleControl:
		body = &VehicleControlMessage{}
	case MsgIDCircularAreaSet:
		body = &CircularAreaSetMessage{}
	case MsgIDRectAreaSet:
		body = &RectAreaSetMessage{}
	case MsgIDDriverID:
		body = &DriverIDMessage{}
	case MsgIDCanData:
		body = &CanDataMessage{}
	case MsgIDTerminalPropQuery:
		body = &TerminalPropQueryMessage{}
	case MsgIDTerminalUpgrade:
		body = &TerminalUpgradeMessage{}
	case MsgIDMultimedia:
		body = &MultimediaMessage{}
	case MsgIDPhotoCommand:
		body = &PhotoCommandMessage{}
	case MsgIDEventSet:
		body = &EventSetMessage{}
	case MsgIDEventResp:
		body = &EventRespMessage{}
	case MsgIDAlarm:
		body = &AlarmMessage{}
	case MsgIDOverspeedAlarm:
		body = &OverspeedAlarmMessage{}
	case MsgIDFatigueDriveAlarm:
		body = &FatigueDriveAlarmMessage{}
	case MsgIDTerminalPropResp:
		body = &TerminalPropRespMessage{}
	case MsgIDTerminalUpgradeResp:
		body = &TerminalUpgradeRespMessage{}
	case MsgIDLocationQueryResp:
		body = &LocationQueryRespMessage{}
	case MsgIDTempLocationTrackResp:
		body = &TempLocationTrackRespMessage{}
	case MsgIDQuestionResp:
		body = &QuestionRespMessage{}
	case MsgIDInfoMenuResp:
		body = &InfoMenuRespMessage{}
	case MsgIDSMSForwardResp:
		body = &SMSForwardRespMessage{}
	case MsgIDMultimediaUpload:
		body = &MultimediaUploadMessage{}
	case MsgIDStorageMediaSearch:
		body = &StorageMediaSearchMessage{}
	case MsgIDStorageMediaUpload:
		body = &StorageMediaUploadMessage{}
	case MsgIDPhotoCommandResp:
		body = &PhotoCommandRespMessage{}
	case MsgIDAlarmAttachment:
		body = &AlarmAttachmentMessage{}
	case MsgIDAlarmAttachmentResp:
		body = &AlarmAttachmentRespMessage{}
	case MsgIDRSAPublicKey:
		body = &RSAPublicKeyMessage{}
	case MsgIDRSADistribute:
		body = &RSADistributeMessage{}
	case MsgIDBillOperate:
		body = &BillOperateMessage{}
	case MsgIDElectronicWaybill:
		body = &ElectronicWaybillMessage{}
	case MsgIDPolygonAreaSet:
		body = &PolygonAreaSetMessage{}
	case MsgIDRouteSet:
		body = &RouteSetMessage{}
	default:
		body = &RawMessage{ID: msgID, Data: data}
	}

	if err := body.Unmarshal(data); err != nil {
		return nil, fmt.Errorf("unmarshal msg 0x%04X: %w", msgID, err)
	}

	return body, nil
}

// EncodeBody 编码消息体
func (c *JT808Codec) EncodeBody(body types.MessageBody) ([]byte, error) {
	return body.Marshal()
}

// VerifyChecksum 校验校验码
func (c *JT808Codec) VerifyChecksum(data []byte) bool {
	if len(data) < 2 {
		return false
	}
	var xor byte
	for i := 0; i < len(data)-1; i++ {
		xor ^= data[i]
	}
	return xor == data[len(data)-1]
}

// Encode 完整编码（头+体+校验码+转义+分隔符）
func (c *JT808Codec) Encode(header *types.MessageHeader, body types.MessageBody) ([]byte, error) {
	headerBytes, err := c.EncodeHeader(header)
	if err != nil {
		return nil, fmt.Errorf("encode header: %w", err)
	}

	bodyBytes, err := c.EncodeBody(body)
	if err != nil {
		return nil, fmt.Errorf("encode body: %w", err)
	}

	// 更新消息体属性中的长度
	bodyAttr := uint16(len(bodyBytes)) & 0x03FF
	bodyAttr |= (uint16(header.EncryptMethod) & 0x07) << 10
	if header.Version2019 {
		bodyAttr |= 0x8000
	}
	if header.HasPack {
		bodyAttr |= 0x2000
	}
	binary.BigEndian.PutUint16(headerBytes[2:4], bodyAttr)

	// 组装：头 + 体 + 校验码
	raw := make([]byte, 0, len(headerBytes)+len(bodyBytes)+1)
	raw = append(raw, headerBytes...)
	raw = append(raw, bodyBytes...)
	checksum := CalcChecksum(raw)
	raw = append(raw, checksum)

	// 转义
	escaped := Escape(raw)

	// 添加分隔符
	result := WrapWithDelimiter(escaped)

	return result, nil
}

// Decode 完整解码（分隔符→反转义→校验→头→体）
func (c *JT808Codec) Decode(frame []byte) (*types.Message, error) {
	// 去除首尾分隔符
	striped := StripDelimiter(frame)

	// 反转义
	raw, err := Unescape(striped)
	if err != nil {
		return nil, fmt.Errorf("unescape: %w", err)
	}

	// 校验码校验
	if !c.VerifyChecksum(raw) {
		return nil, ErrInvalidChecksum
	}

	// 去除校验码
	data := raw[:len(raw)-1]

	// 解析消息头
	header, headerLen, err := c.ParseHeader(data)
	if err != nil {
		return nil, fmt.Errorf("parse header: %w", err)
	}

	// 解析消息体
	bodyData := data[headerLen:]
	body, err := c.ParseBody(header.MsgID, bodyData)
	if err != nil {
		return nil, fmt.Errorf("parse body: %w", err)
	}

	return &types.Message{
		Header: *header,
		Body:   body,
		Raw:    frame,
	}, nil
}

// MsgName 消息名称
func MsgName(msgID uint16) string {
	for _, m := range AllMessages {
		if m.ID == msgID {
			return m.Name
		}
	}
	return fmt.Sprintf("未知消息(0x%04X)", msgID)
}

// MsgDirection 消息方向
func MsgDirection(msgID uint16) types.Direction {
	if msgID&0x8000 != 0 {
		return types.DirectionDown
	}
	return types.DirectionUp
}
