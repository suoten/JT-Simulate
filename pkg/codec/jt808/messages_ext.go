package jt808

import (
	"encoding/binary"
	"fmt"
)

// ============================== 0x0107 终端属性应答 ==============================

type TerminalPropRespMessage struct {
	TerminalType   uint16 // 终端类型
	ManufacturerID string // 制造商ID(5字节)
	TerminalModel  string // 终端型号(20字节)
	TerminalID     string // 终端ID(7字节)
	ICCID          string // SIM卡ICCID(10字节)
	HardwareVer    string // 硬件版本号
	FirmwareVer    string // 固件版本号
	GPSModule      byte   // GNSS模块属性
	COMModule      byte   // 通信模块属性
}

func (m *TerminalPropRespMessage) MsgID() uint16 { return MsgIDTerminalPropResp }

func (m *TerminalPropRespMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 0, 80)
	typeBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(typeBytes, m.TerminalType)
	buf = append(buf, typeBytes...)

	manu := []byte(m.ManufacturerID)
	for len(manu) < 5 {
		manu = append(manu, 0)
	}
	buf = append(buf, manu[:5]...)

	model := []byte(m.TerminalModel)
	for len(model) < 20 {
		model = append(model, 0)
	}
	buf = append(buf, model[:20]...)

	tid := []byte(m.TerminalID)
	for len(tid) < 7 {
		tid = append(tid, 0)
	}
	buf = append(buf, tid[:7]...)

	iccid := []byte(m.ICCID)
	for len(iccid) < 10 {
		iccid = append(iccid, 0)
	}
	buf = append(buf, iccid[:10]...)

	buf = append(buf, m.GPSModule)
	buf = append(buf, m.COMModule)

	hwVerLen := byte(len(m.HardwareVer))
	buf = append(buf, hwVerLen)
	buf = append(buf, []byte(m.HardwareVer)...)

	fwVerLen := byte(len(m.FirmwareVer))
	buf = append(buf, fwVerLen)
	buf = append(buf, []byte(m.FirmwareVer)...)

	return buf, nil
}

func (m *TerminalPropRespMessage) Unmarshal(data []byte) error {
	if len(data) < 47 {
		return ErrDataTooShort
	}
	m.TerminalType = binary.BigEndian.Uint16(data[0:2])
	m.ManufacturerID = trimNull(data[2:7])
	m.TerminalModel = trimNull(data[7:27])
	m.TerminalID = trimNull(data[27:34])
	m.ICCID = trimNull(data[34:44])
	m.GPSModule = data[44]
	m.COMModule = data[45]
	offset := 46
	if offset < len(data) {
		hwLen := int(data[offset])
		offset++
		if offset+hwLen <= len(data) {
			m.HardwareVer = string(data[offset : offset+hwLen])
			offset += hwLen
		}
	}
	if offset < len(data) {
		fwLen := int(data[offset])
		offset++
		if offset+fwLen <= len(data) {
			m.FirmwareVer = string(data[offset : offset+fwLen])
		}
	}
	return nil
}

// ============================== 0x0108 终端升级应答 ==============================

type TerminalUpgradeRespMessage struct {
	UpgradeType byte   // 升级类型
	Result      byte   // 0=成功 1=失败 2=取消
	UpgradeMsg  string // 升级结果消息
}

func (m *TerminalUpgradeRespMessage) MsgID() uint16 { return MsgIDTerminalUpgradeResp }

func (m *TerminalUpgradeRespMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 0, 20)
	buf = append(buf, m.UpgradeType)
	buf = append(buf, m.Result)
	buf = append(buf, []byte(m.UpgradeMsg)...)
	return buf, nil
}

func (m *TerminalUpgradeRespMessage) Unmarshal(data []byte) error {
	if len(data) < 2 {
		return ErrDataTooShort
	}
	m.UpgradeType = data[0]
	m.Result = data[1]
	if len(data) > 2 {
		m.UpgradeMsg = string(data[2:])
	}
	return nil
}

// ============================== 0x0201 位置查询应答 ==============================

type LocationQueryRespMessage struct {
	RespSeqNum uint16 // 应答流水号(对应位置查询消息的流水号)
	LocationMessage      // 嵌入位置信息
}

func (m *LocationQueryRespMessage) MsgID() uint16 { return MsgIDLocationQueryResp }

func (m *LocationQueryRespMessage) Marshal() ([]byte, error) {
	seq := make([]byte, 2)
	binary.BigEndian.PutUint16(seq, m.RespSeqNum)
	locData, err := m.LocationMessage.Marshal()
	if err != nil {
		return nil, err
	}
	return append(seq, locData...), nil
}

func (m *LocationQueryRespMessage) Unmarshal(data []byte) error {
	if len(data) < 2 {
		return ErrDataTooShort
	}
	m.RespSeqNum = binary.BigEndian.Uint16(data[0:2])
	return m.LocationMessage.Unmarshal(data[2:])
}

// ============================== 0x0202 临时位置跟踪应答 ==============================

type TempLocationTrackRespMessage struct {
	RespSeqNum uint16 // 应答流水号
	Result     byte   // 0=成功 1=失败
}

func (m *TempLocationTrackRespMessage) MsgID() uint16 { return MsgIDTempLocationTrackResp }

func (m *TempLocationTrackRespMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 3)
	binary.BigEndian.PutUint16(buf[0:2], m.RespSeqNum)
	buf[2] = m.Result
	return buf, nil
}

func (m *TempLocationTrackRespMessage) Unmarshal(data []byte) error {
	if len(data) < 3 {
		return ErrDataTooShort
	}
	m.RespSeqNum = binary.BigEndian.Uint16(data[0:2])
	m.Result = data[2]
	return nil
}

// ============================== 0x0302 提问应答 ==============================

type QuestionRespMessage struct {
	RespSeqNum uint16 // 应答流水号
	AnswerID   byte   // 答案ID
}

func (m *QuestionRespMessage) MsgID() uint16 { return MsgIDQuestionResp }

func (m *QuestionRespMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 3)
	binary.BigEndian.PutUint16(buf[0:2], m.RespSeqNum)
	buf[2] = m.AnswerID
	return buf, nil
}

func (m *QuestionRespMessage) Unmarshal(data []byte) error {
	if len(data) < 3 {
		return ErrDataTooShort
	}
	m.RespSeqNum = binary.BigEndian.Uint16(data[0:2])
	m.AnswerID = data[2]
	return nil
}

// ============================== 0x0700 信息菜单应答 ==============================

type InfoMenuRespMessage struct {
	RespSeqNum uint16 // 应答流水号
	MenuType   byte   // 菜单类型
}

func (m *InfoMenuRespMessage) MsgID() uint16 { return MsgIDInfoMenuResp }

func (m *InfoMenuRespMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 3)
	binary.BigEndian.PutUint16(buf[0:2], m.RespSeqNum)
	buf[2] = m.MenuType
	return buf, nil
}

func (m *InfoMenuRespMessage) Unmarshal(data []byte) error {
	if len(data) < 3 {
		return ErrDataTooShort
	}
	m.RespSeqNum = binary.BigEndian.Uint16(data[0:2])
	m.MenuType = data[2]
	return nil
}

// ============================== 0x0703 短信转发应答 ==============================

type SMSForwardRespMessage struct {
	RespSeqNum uint16 // 应答流水号
	Result     byte   // 0=成功 1=失败
}

func (m *SMSForwardRespMessage) MsgID() uint16 { return MsgIDSMSForwardResp }

func (m *SMSForwardRespMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 3)
	binary.BigEndian.PutUint16(buf[0:2], m.RespSeqNum)
	buf[2] = m.Result
	return buf, nil
}

func (m *SMSForwardRespMessage) Unmarshal(data []byte) error {
	if len(data) < 3 {
		return ErrDataTooShort
	}
	m.RespSeqNum = binary.BigEndian.Uint16(data[0:2])
	m.Result = data[2]
	return nil
}

// ============================== 0x0802 多媒体数据上传 ==============================

type MultimediaUploadMessage struct {
	MultimediaID   uint32 // 多媒体ID
	MultimediaType byte   // 多媒体类型
	Format         byte   // 编码格式
	PlayTime       uint16 // 播放时长(秒)
	PackageSize    byte   // 每包大小
	TotalPackages  uint16 // 总包数
	Offset         uint16 // 偏移量
	Data           []byte // 多媒体数据
}

func (m *MultimediaUploadMessage) MsgID() uint16 { return MsgIDMultimediaUpload }

func (m *MultimediaUploadMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 0, 20+len(m.Data))
	idBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(idBytes, m.MultimediaID)
	buf = append(buf, idBytes...)
	buf = append(buf, m.MultimediaType)
	buf = append(buf, m.Format)
	ptBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(ptBytes, m.PlayTime)
	buf = append(buf, ptBytes...)
	buf = append(buf, m.PackageSize)
	tpBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(tpBytes, m.TotalPackages)
	buf = append(buf, tpBytes...)
	ofBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(ofBytes, m.Offset)
	buf = append(buf, ofBytes...)
	buf = append(buf, m.Data...)
	return buf, nil
}

func (m *MultimediaUploadMessage) Unmarshal(data []byte) error {
	if len(data) < 13 {
		return ErrDataTooShort
	}
	m.MultimediaID = binary.BigEndian.Uint32(data[0:4])
	m.MultimediaType = data[4]
	m.Format = data[5]
	m.PlayTime = binary.BigEndian.Uint16(data[6:8])
	m.PackageSize = data[8]
	m.TotalPackages = binary.BigEndian.Uint16(data[9:11])
	m.Offset = binary.BigEndian.Uint16(data[11:13])
	if len(data) > 13 {
		m.Data = make([]byte, len(data)-13)
		copy(m.Data, data[13:])
	}
	return nil
}

// ============================== 0x0803 存储多媒体检索 ==============================

type StorageMediaSearchMessage struct {
	SeqNum      uint16 // 流水号
	LogicalCh   byte   // 逻辑通道号
	StartTime   string // 起始时间 YYMMDDHHmmss
	EndTime     string // 结束时间
	AlarmFlag   byte   // 报警标志
	MediaType   byte   // 媒体类型
	StreamType  byte   // 码流类型
}

func (m *StorageMediaSearchMessage) MsgID() uint16 { return MsgIDStorageMediaSearch }

func (m *StorageMediaSearchMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 0, 20)
	seq := make([]byte, 2)
	binary.BigEndian.PutUint16(seq, m.SeqNum)
	buf = append(buf, seq...)
	buf = append(buf, m.LogicalCh)
	for _, t := range []string{m.StartTime, m.EndTime} {
		for len(t) < 12 {
			t = "0" + t
		}
		for i := 0; i < 6; i++ {
			buf = append(buf, ((t[i*2]-'0')<<4)|(t[i*2+1]-'0'))
		}
	}
	buf = append(buf, m.AlarmFlag, m.MediaType, m.StreamType)
	return buf, nil
}

func (m *StorageMediaSearchMessage) Unmarshal(data []byte) error {
	if len(data) < 17 {
		return ErrDataTooShort
	}
	m.SeqNum = binary.BigEndian.Uint16(data[0:2])
	m.LogicalCh = data[2]
	decBCD := func(off int) string {
		s := make([]byte, 0, 12)
		for i := 0; i < 6; i++ {
			b := data[off+i]
			s = append(s, (b>>4)+'0', (b&0x0F)+'0')
		}
		return string(s)
	}
	m.StartTime = decBCD(3)
	m.EndTime = decBCD(9)
	m.AlarmFlag = data[15]
	m.MediaType = data[16]
	if len(data) > 17 {
		m.StreamType = data[17]
	}
	return nil
}

// ============================== 0x0804 存储多媒体上传 ==============================

type StorageMediaUploadMessage struct {
	RespSeqNum  uint16 // 应答流水号
	Result      byte   // 0=成功 1=失败
	MultimediaID uint32 // 多媒体ID
	PackageSize  byte   // 每包大小
	Offset       uint16 // 起始偏移
}

func (m *StorageMediaUploadMessage) MsgID() uint16 { return MsgIDStorageMediaUpload }

func (m *StorageMediaUploadMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 10)
	binary.BigEndian.PutUint16(buf[0:2], m.RespSeqNum)
	buf[2] = m.Result
	binary.BigEndian.PutUint32(buf[3:7], m.MultimediaID)
	buf[7] = m.PackageSize
	binary.BigEndian.PutUint16(buf[8:10], m.Offset)
	return buf, nil
}

func (m *StorageMediaUploadMessage) Unmarshal(data []byte) error {
	if len(data) < 10 {
		return ErrDataTooShort
	}
	m.RespSeqNum = binary.BigEndian.Uint16(data[0:2])
	m.Result = data[2]
	m.MultimediaID = binary.BigEndian.Uint32(data[3:7])
	m.PackageSize = data[7]
	m.Offset = binary.BigEndian.Uint16(data[8:10])
	return nil
}

// ============================== 0x0805 摄像头拍摄应答 ==============================

type PhotoCommandRespMessage struct {
	RespSeqNum   uint16 // 应答流水号
	Result       byte   // 0=成功 1=失败
	MultimediaID uint32 // 多媒体ID
	ChannelID    byte   // 通道号
	PackageSize  uint16 // 每包大小
}

func (m *PhotoCommandRespMessage) MsgID() uint16 { return MsgIDPhotoCommandResp }

func (m *PhotoCommandRespMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 10)
	binary.BigEndian.PutUint16(buf[0:2], m.RespSeqNum)
	buf[2] = m.Result
	binary.BigEndian.PutUint32(buf[3:7], m.MultimediaID)
	buf[7] = m.ChannelID
	binary.BigEndian.PutUint16(buf[8:10], m.PackageSize)
	return buf, nil
}

func (m *PhotoCommandRespMessage) Unmarshal(data []byte) error {
	if len(data) < 10 {
		return ErrDataTooShort
	}
	m.RespSeqNum = binary.BigEndian.Uint16(data[0:2])
	m.Result = data[2]
	m.MultimediaID = binary.BigEndian.Uint32(data[3:7])
	m.ChannelID = data[7]
	m.PackageSize = binary.BigEndian.Uint16(data[8:10])
	return nil
}

// ============================== 0x0901 报警附件 ==============================

type AlarmAttachmentMessage struct {
	VehicleColor  byte   // 车牌颜色
	VehiclePlate  string // 车牌号码
	AlarmID       uint16 // 报警ID
	AttachmentTime string // 附件时间 YYMMDDHHmmss
	AttachmentLen uint16 // 附件数据长度
	AttachmentData []byte // 附件数据
}

func (m *AlarmAttachmentMessage) MsgID() uint16 { return MsgIDAlarmAttachment }

func (m *AlarmAttachmentMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 0, 40+len(m.AttachmentData))
	buf = append(buf, m.VehicleColor)
	plate := []byte(m.VehiclePlate)
	for len(plate) < 21 {
		plate = append(plate, 0)
	}
	buf = append(buf, plate[:21]...)
	idBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(idBytes, m.AlarmID)
	buf = append(buf, idBytes...)
	t := m.AttachmentTime
	for len(t) < 12 {
		t = "0" + t
	}
	for i := 0; i < 6; i++ {
		buf = append(buf, ((t[i*2]-'0')<<4)|(t[i*2+1]-'0'))
	}
	lenBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(lenBytes, m.AttachmentLen)
	buf = append(buf, lenBytes...)
	buf = append(buf, m.AttachmentData...)
	return buf, nil
}

func (m *AlarmAttachmentMessage) Unmarshal(data []byte) error {
	if len(data) < 33 {
		return ErrDataTooShort
	}
	m.VehicleColor = data[0]
	plate := data[1:22]
	for len(plate) > 0 && plate[len(plate)-1] == 0 {
		plate = plate[:len(plate)-1]
	}
	m.VehiclePlate = string(plate)
	m.AlarmID = binary.BigEndian.Uint16(data[22:24])
	t := make([]byte, 0, 12)
	for i := 0; i < 6; i++ {
		b := data[24+i]
		t = append(t, (b>>4)+'0', (b&0x0F)+'0')
	}
	m.AttachmentTime = string(t)
	m.AttachmentLen = binary.BigEndian.Uint16(data[30:32])
	if len(data) > 32 {
		m.AttachmentData = make([]byte, len(data)-32)
		copy(m.AttachmentData, data[32:])
	}
	return nil
}

// ============================== 0x9001 报警附件应答 ==============================

type AlarmAttachmentRespMessage struct {
	RespSeqNum uint16 // 应答流水号
	Result     byte   // 0=成功 1=失败
}

func (m *AlarmAttachmentRespMessage) MsgID() uint16 { return MsgIDAlarmAttachmentResp }

func (m *AlarmAttachmentRespMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 3)
	binary.BigEndian.PutUint16(buf[0:2], m.RespSeqNum)
	buf[2] = m.Result
	return buf, nil
}

func (m *AlarmAttachmentRespMessage) Unmarshal(data []byte) error {
	if len(data) < 3 {
		return ErrDataTooShort
	}
	m.RespSeqNum = binary.BigEndian.Uint16(data[0:2])
	m.Result = data[2]
	return nil
}

// ============================== 0x0A00 终端RSA公钥 ==============================

type RSAPublicKeyMessage struct {
	EModule    []byte // RSA模数(e)
	Exponent   []byte // RSA指数(公钥指数)
}

func (m *RSAPublicKeyMessage) MsgID() uint16 { return MsgIDRSAPublicKey }

func (m *RSAPublicKeyMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 0, 300)
	buf = append(buf, byte(len(m.EModule)))
	buf = append(buf, m.EModule...)
	buf = append(buf, byte(len(m.Exponent)))
	buf = append(buf, m.Exponent...)
	return buf, nil
}

func (m *RSAPublicKeyMessage) Unmarshal(data []byte) error {
	if len(data) < 2 {
		return ErrDataTooShort
	}
	eLen := int(data[0])
	if 1+eLen >= len(data) {
		return ErrDataTooShort
	}
	m.EModule = make([]byte, eLen)
	copy(m.EModule, data[1:1+eLen])
	expLen := int(data[1+eLen])
	if 1+eLen+1+expLen > len(data) {
		return ErrDataTooShort
	}
	m.Exponent = make([]byte, expLen)
	copy(m.Exponent, data[1+eLen+1:])
	return nil
}

// ============================== 0x8A00 平台RSA公钥下发 ==============================

type RSADistributeMessage struct {
	EModule    []byte // RSA模数
	Exponent   []byte // RSA指数
}

func (m *RSADistributeMessage) MsgID() uint16 { return MsgIDRSADistribute }

func (m *RSADistributeMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 0, 300)
	buf = append(buf, byte(len(m.EModule)))
	buf = append(buf, m.EModule...)
	buf = append(buf, byte(len(m.Exponent)))
	buf = append(buf, m.Exponent...)
	return buf, nil
}

func (m *RSADistributeMessage) Unmarshal(data []byte) error {
	if len(data) < 2 {
		return ErrDataTooShort
	}
	eLen := int(data[0])
	if 1+eLen >= len(data) {
		return ErrDataTooShort
	}
	m.EModule = make([]byte, eLen)
	copy(m.EModule, data[1:1+eLen])
	expLen := int(data[1+eLen])
	if 1+eLen+1+expLen > len(data) {
		return ErrDataTooShort
	}
	m.Exponent = make([]byte, expLen)
	copy(m.Exponent, data[1+eLen+1:])
	return nil
}

// ============================== 0x0B00 计价器操作 ==============================

type BillOperateMessage struct {
	OperateType byte   // 操作类型
	OperateData []byte // 操作数据
}

func (m *BillOperateMessage) MsgID() uint16 { return MsgIDBillOperate }

func (m *BillOperateMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 0, 20)
	buf = append(buf, m.OperateType)
	buf = append(buf, m.OperateData...)
	return buf, nil
}

func (m *BillOperateMessage) Unmarshal(data []byte) error {
	if len(data) < 1 {
		return ErrDataTooShort
	}
	m.OperateType = data[0]
	if len(data) > 1 {
		m.OperateData = make([]byte, len(data)-1)
		copy(m.OperateData, data[1:])
	}
	return nil
}

// ============================== 0x0701 电子运单 ==============================

type ElectronicWaybillMessage struct {
	WaybillData []byte // 电子运单数据(变长，JSON或XML)
}

func (m *ElectronicWaybillMessage) MsgID() uint16 { return MsgIDElectronicWaybill }

func (m *ElectronicWaybillMessage) Marshal() ([]byte, error) {
	return m.WaybillData, nil
}

func (m *ElectronicWaybillMessage) Unmarshal(data []byte) error {
	m.WaybillData = make([]byte, len(data))
	copy(m.WaybillData, data)
	return nil
}

// ============================== 0x8604 多边形区域设置 ==============================

type PolygonAreaSetMessage struct {
	AreaID    uint32   // 区域ID
	Attr      uint16   // 区域属性
	StartTime string   // 起始时间 HHMMSS
	EndTime   string   // 结束时间
	SpeedLimit uint16  // 最高速度
	Points    []PolygonPoint // 顶点
}

type PolygonPoint struct {
	Lat float64
	Lon float64
}

func (m *PolygonAreaSetMessage) MsgID() uint16 { return MsgIDPolygonAreaSet }

func (m *PolygonAreaSetMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 0, 20+len(m.Points)*8)
	idBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(idBytes, m.AreaID)
	buf = append(buf, idBytes...)

	attrBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(attrBytes, m.Attr)
	buf = append(buf, attrBytes...)

	ptCount := make([]byte, 2)
	binary.BigEndian.PutUint16(ptCount, uint16(len(m.Points)))
	buf = append(buf, ptCount...)

	for _, pt := range m.Points {
		latBytes := make([]byte, 4)
		binary.BigEndian.PutUint32(latBytes, uint32(pt.Lat*CoordScaleFactor))
		buf = append(buf, latBytes...)
		lonBytes := make([]byte, 4)
		binary.BigEndian.PutUint32(lonBytes, uint32(pt.Lon*CoordScaleFactor))
		buf = append(buf, lonBytes...)
	}

	if m.Attr&0x02 != 0 {
		startBCD, _ := StringToBCD(m.StartTime, 3)
		buf = append(buf, startBCD...)
		endBCD, _ := StringToBCD(m.EndTime, 3)
		buf = append(buf, endBCD...)
		slBytes := make([]byte, 2)
		binary.BigEndian.PutUint16(slBytes, m.SpeedLimit)
		buf = append(buf, slBytes...)
	}

	return buf, nil
}

func (m *PolygonAreaSetMessage) Unmarshal(data []byte) error {
	if len(data) < 8 {
		return ErrDataTooShort
	}
	m.AreaID = binary.BigEndian.Uint32(data[0:4])
	m.Attr = binary.BigEndian.Uint16(data[4:6])
	ptCount := int(binary.BigEndian.Uint16(data[6:8]))
	offset := 8
	for i := 0; i < ptCount; i++ {
		if offset+8 > len(data) {
			return fmt.Errorf("polygon point %d out of range", i)
		}
		var pt PolygonPoint
		pt.Lat = float64(binary.BigEndian.Uint32(data[offset:offset+4])) / CoordScaleFactor
		pt.Lon = float64(binary.BigEndian.Uint32(data[offset+4:offset+8])) / CoordScaleFactor
		m.Points = append(m.Points, pt)
		offset += 8
	}
	if m.Attr&0x02 != 0 && offset+8 <= len(data) {
		m.StartTime = BCDToStringSafe(data[offset : offset+3])
		m.EndTime = BCDToStringSafe(data[offset+3 : offset+6])
		m.SpeedLimit = binary.BigEndian.Uint16(data[offset+6 : offset+8])
	}
	return nil
}

// ============================== 0x8606 路线设置 ==============================

type RouteSetMessage struct {
	RouteID    uint32   // 路线ID
	Attr       uint16   // 路线属性
	StartTime  string   // 起始时间 HHMMSS
	EndTime    string   // 结束时间
	Points     []RoutePoint // 苐点
}

type RoutePoint struct {
	PointID    uint32  // 路段ID
	Lat        float64 // 纬度
	Lon        float64 // 经度
	Width      byte    // 路段宽度
	Attr       byte    // 路段属性
	MaxSpeed   uint16  // 最高速度
	MaxDuration uint16 // 最长行驶时间
}

func (m *RouteSetMessage) MsgID() uint16 { return MsgIDRouteSet }

func (m *RouteSetMessage) Marshal() ([]byte, error) {
	buf := make([]byte, 0, 20+len(m.Points)*18)
	idBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(idBytes, m.RouteID)
	buf = append(buf, idBytes...)

	attrBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(attrBytes, m.Attr)
	buf = append(buf, attrBytes...)

	if m.Attr&0x02 != 0 {
		startBCD, _ := StringToBCD(m.StartTime, 3)
		buf = append(buf, startBCD...)
		endBCD, _ := StringToBCD(m.EndTime, 3)
		buf = append(buf, endBCD...)
	}

	ptCount := make([]byte, 2)
	binary.BigEndian.PutUint16(ptCount, uint16(len(m.Points)))
	buf = append(buf, ptCount...)

	for _, pt := range m.Points {
		pid := make([]byte, 4)
		binary.BigEndian.PutUint32(pid, pt.PointID)
		buf = append(buf, pid...)

		latBytes := make([]byte, 4)
		binary.BigEndian.PutUint32(latBytes, uint32(pt.Lat*CoordScaleFactor))
		buf = append(buf, latBytes...)

		lonBytes := make([]byte, 4)
		binary.BigEndian.PutUint32(lonBytes, uint32(pt.Lon*CoordScaleFactor))
		buf = append(buf, lonBytes...)

		buf = append(buf, pt.Width)
		buf = append(buf, pt.Attr)

		spBytes := make([]byte, 2)
		binary.BigEndian.PutUint16(spBytes, pt.MaxSpeed)
		buf = append(buf, spBytes...)

		durBytes := make([]byte, 2)
		binary.BigEndian.PutUint16(durBytes, pt.MaxDuration)
		buf = append(buf, durBytes...)
	}

	return buf, nil
}

func (m *RouteSetMessage) Unmarshal(data []byte) error {
	if len(data) < 8 {
		return ErrDataTooShort
	}
	m.RouteID = binary.BigEndian.Uint32(data[0:4])
	m.Attr = binary.BigEndian.Uint16(data[4:6])
	offset := 6
	if m.Attr&0x02 != 0 {
		if offset+6 > len(data) {
			return ErrDataTooShort
		}
		m.StartTime = BCDToStringSafe(data[offset : offset+3])
		m.EndTime = BCDToStringSafe(data[offset+3 : offset+6])
		offset += 6
	}
	if offset+2 > len(data) {
		return ErrDataTooShort
	}
	ptCount := int(binary.BigEndian.Uint16(data[offset : offset+2]))
	offset += 2
	for i := 0; i < ptCount; i++ {
		if offset+18 > len(data) {
			return fmt.Errorf("route point %d out of range", i)
		}
		var pt RoutePoint
		pt.PointID = binary.BigEndian.Uint32(data[offset : offset+4])
		pt.Lat = float64(binary.BigEndian.Uint32(data[offset+4:offset+8])) / CoordScaleFactor
		pt.Lon = float64(binary.BigEndian.Uint32(data[offset+8:offset+12])) / CoordScaleFactor
		pt.Width = data[offset+12]
		pt.Attr = data[offset+13]
		pt.MaxSpeed = binary.BigEndian.Uint16(data[offset+14 : offset+16])
		pt.MaxDuration = binary.BigEndian.Uint16(data[offset+16 : offset+18])
		m.Points = append(m.Points, pt)
		offset += 18
	}
	return nil
}
