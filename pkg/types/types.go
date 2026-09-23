package types

// ProtocolType 协议类型
type ProtocolType string

const (
	ProtocolJT808   ProtocolType = "jt808"
	ProtocolJT809   ProtocolType = "jt809"
	ProtocolJT1078  ProtocolType = "jt1078"
	ProtocolJT905   ProtocolType = "jt905"
	ProtocolJT1045  ProtocolType = "jt1045"
	ProtocolJT1253  ProtocolType = "jt1253"
	ProtocolGBT32960 ProtocolType = "gbt32960"
)

// MessageHeader 消息头（通用）
type MessageHeader struct {
	MsgID         uint16
	BodyAttr      uint16
	Phone         string
	SeqNum        uint16
	PackTotal     uint16
	PackIndex     uint16
	BodyLen       int
	HasPack       bool
	EncryptMethod uint8
	PlateColor    byte
	Version2019   bool
	ProtocolVer   byte
}

// MessageBody 消息体接口
type MessageBody interface {
	MsgID() uint16
	Marshal() ([]byte, error)
	Unmarshal(data []byte) error
}

// Message 完整消息
type Message struct {
	Header MessageHeader
	Body   MessageBody
	Raw    []byte
}

// Codec 编解码接口
type Codec interface {
	ProtocolType() ProtocolType
	ParseHeader(data []byte) (*MessageHeader, int, error)
	EncodeHeader(header *MessageHeader) ([]byte, error)
	ParseBody(msgID uint16, data []byte) (MessageBody, error)
	EncodeBody(body MessageBody) ([]byte, error)
	VerifyChecksum(data []byte) bool
}

// Direction 消息方向
type Direction string

const (
	DirectionUp   Direction = "up"   // 终端→平台
	DirectionDown Direction = "down" // 平台→终端
)

// MessageInfo 消息信息（用于前端展示）
type MessageInfo struct {
	Protocol  ProtocolType `json:"protocol"`
	MsgID     uint16       `json:"msg_id"`
	MsgName   string       `json:"msg_name"`
	Direction Direction    `json:"direction"`
	Hex       string       `json:"hex"`
	Parsed    interface{}  `json:"parsed"`
	Issues    []Issue      `json:"issues,omitempty"`
}

// Issue 问题信息
type Issue struct {
	Level   string `json:"level"` // error, warning, info
	Field   string `json:"field"`
	Message string `json:"message"`
}
