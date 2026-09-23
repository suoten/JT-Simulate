package jt808

import "errors"

// ==================== JT/T 808-2019 消息ID常量 ====================
const (
	// 通用
	MsgIDTerminalGeneralResp   uint16 = 0x0001 // 终端通用应答
	MsgIDPlatformGeneralResp   uint16 = 0x8001 // 平台通用应答
	MsgIDHeartbeat             uint16 = 0x0002 // 终端心跳
	MsgIDTerminalCancel        uint16 = 0x0003 // 终端注销
	MsgIDTerminalCancelResp    uint16 = 0x8003 // 终端注销应答

	// 注册与鉴权
	MsgIDRegister              uint16 = 0x0100 // 终端注册
	MsgIDRegisterResp          uint16 = 0x8100 // 终端注册应答
	MsgIDAuth                  uint16 = 0x0102 // 终端鉴权

	// 参数设置/查询
	MsgIDCommand               uint16 = 0x8103 // 设置参数
	MsgIDCommandResp           uint16 = 0x0103 // 设置参数应答
	MsgIDParamQuery            uint16 = 0x8104 // 查询参数
	MsgIDParamResp             uint16 = 0x0104 // 查询参数应答
	MsgIDParamQuerySpecific    uint16 = 0x8106 // 查询指定参数
	MsgIDTerminalCtrl          uint16 = 0x8105 // 终端控制
	MsgIDTerminalPropQuery     uint16 = 0x8107 // 终端属性查询
	MsgIDTerminalPropResp      uint16 = 0x0107 // 终端属性应答
	MsgIDTerminalUpgrade       uint16 = 0x8108 // 终端升级
	MsgIDTerminalUpgradeResp   uint16 = 0x0108 // 终端升级应答

	// 位置相关
	MsgIDLocation              uint16 = 0x0200 // 位置上报
	MsgIDLocationQuery         uint16 = 0x8201 // 位置查询
	MsgIDLocationQueryResp     uint16 = 0x0201 // 位置查询应答
	MsgIDTempLocationTrack     uint16 = 0x8202 // 临时位置跟踪
	MsgIDTempLocationTrackResp uint16 = 0x0202 // 临时位置跟踪应答
	MsgIDManualAlarmConfirm    uint16 = 0x8203 // 人工确认报警
	MsgIDLocationBatch         uint16 = 0x0704 // 批量位置上报

	// 报警
	MsgIDAlarm                 uint16 = 0x0900 // 报警
	MsgIDAlarmAck              uint16 = 0x8900 // 报警确认
	MsgIDAlarmAttachment       uint16 = 0x0901 // 报警附件
	MsgIDAlarmAttachmentResp   uint16 = 0x9001 // 报警附件应答

	// 文本与事件
	MsgIDTextSend              uint16 = 0x8300 // 文本下发
	MsgIDEventSet              uint16 = 0x8301 // 事件设置
	MsgIDEventResp             uint16 = 0x0301 // 事件应答
	MsgIDEventDel              uint16 = 0x8302 // 事件删除 (2019: 提问下发改为事件删除)
	MsgIDQuestionDown          uint16 = 0x8302 // 提问下发 (2011/2013)
	MsgIDQuestionResp          uint16 = 0x0302 // 提问应答
	MsgIDInfoDistribute        uint16 = 0x8303 // 信息点下发
	MsgIDInfoService           uint16 = 0x8304 // 信息服务
	MsgIDPhoneBookSet          uint16 = 0x8204 // 电话簿设置

	// 报警设置
	MsgIDOverspeedSet          uint16 = 0x8400 // 超速报警设置
	MsgIDOverspeedAlarm        uint16 = 0x0400 // 超速报警
	MsgIDFatigueDriveSet       uint16 = 0x8401 // 疲劳驾驶设置
	MsgIDFatigueDriveAlarm     uint16 = 0x0401 // 疲劳驾驶报警
	MsgIDAreaRouteAlarmSet     uint16 = 0x8402 // 区域路线报警设置
	MsgIDAreaRouteAlarmDel     uint16 = 0x8403 // 区域路线报警删除
	MsgIDEWaybillSet           uint16 = 0x8404 // 电子运单设置
	MsgIDEWaybillDel           uint16 = 0x8405 // 电子运单删除
	MsgIDEWaybillUpload        uint16 = 0x8406 // 电子运单上报
	MsgIDEWaybillResp          uint16 = 0x8407 // 电子运单应答

	// 车辆控制
	MsgIDVehicleControl        uint16 = 0x8500 // 车辆控制

	// 区域设置
	MsgIDCircularAreaSet       uint16 = 0x8600 // 圆形区域设置
	MsgIDCircularAreaDel       uint16 = 0x8601 // 圆形区域删除
	MsgIDRectAreaSet           uint16 = 0x8602 // 矩形区域设置
	MsgIDRectAreaDel           uint16 = 0x8603 // 矩形区域删除
	MsgIDPolygonAreaSet        uint16 = 0x8604 // 多边形区域设置
	MsgIDPolygonAreaDel        uint16 = 0x8605 // 多边形区域删除
	MsgIDRouteSet              uint16 = 0x8606 // 路线设置
	MsgIDRouteDel              uint16 = 0x8607 // 路线删除
	MsgIDFireAreaSet           uint16 = 0x8608 // 火源区域设置
	MsgIDFireAreaDel           uint16 = 0x8609 // 火源区域删除
	MsgIDFireAreaAlarm         uint16 = 0x0500 // 区域路线报警

	// 信息与通讯
	MsgIDInfoMenuSet           uint16 = 0x8700 // 信息菜单设置
	MsgIDInfoMenuResp          uint16 = 0x0700 // 信息菜单应答
	MsgIDInfoPush              uint16 = 0x8701 // 信息点推送
	MsgIDPhoneCallback         uint16 = 0x8702 // 电话回拨
	MsgIDDriverID              uint16 = 0x0702 // 驾驶员身份
	MsgIDSMSForward            uint16 = 0x8703 // 短信转发
	MsgIDSMSForwardResp        uint16 = 0x0703 // 短信转发应答

	// 多媒体
	MsgIDMultimedia            uint16 = 0x0801 // 多媒体事件
	MsgIDMultimediaUpload      uint16 = 0x0802 // 多媒体数据上传
	MsgIDPhotoCommand         uint16 = 0x8801 // 摄像头立即拍摄
	MsgIDMultimediaUploadCmd  uint16 = 0x8802 // 多媒体数据上传命令
	MsgIDPhotoCommandResp     uint16 = 0x0805 // 摄像头立即拍摄应答
	MsgIDFileUploadCmd        uint16 = 0x8803 // 文件上传命令
	MsgIDAudioRecordCmd       uint16 = 0x8804 // 录音命令
	MsgIDStorageMediaSearch   uint16 = 0x0803 // 存储多媒体检索
	MsgIDStorageMediaUpload   uint16 = 0x0804 // 存储多媒体上传

	// CAN数据与电子运单
	MsgIDCanData              uint16 = 0x0705 // CAN数据
	MsgIDElectronicWaybill    uint16 = 0x0701 // 电子运单

	// RSA与计价器
	MsgIDRSAPublicKey         uint16 = 0x0A00 // 终端RSA公钥
	MsgIDRSADistribute        uint16 = 0x8A00 // 平台RSA公钥下发
	MsgIDBillOperate          uint16 = 0x0B00 // 计价器操作

	// ==================== JT/T 808-2011 独有消息 ====================
	// 2011版特有的部分消息ID与2019相同但字段格式不同

	// ==================== JT/T 808-2013 独有消息 ====================
	// 2013版新增的消息（部分与2019重合）

	// ==================== JT/T 1078 消息（复用808帧格式） ====================
	MsgIDRealtimeAVReq1078   uint16 = 0x9101
	MsgIDRealtimeAVCtrl1078  uint16 = 0x9102
	MsgIDPlaybackReq1078     uint16 = 0x9201
	MsgIDPlaybackResp1078    uint16 = 0x9202
	MsgIDPlaybackCtrl1078    uint16 = 0x9203
	MsgIDPlaybackCtrlAck1078 uint16 = 0x9204
	MsgIDRTPData1078         uint16 = 0x1200
)

// 注册应答结果码（JT/T 808-2019 表5）
const (
	RegResultSuccess         byte = 0 // 成功
	RegResultVehicleExists   byte = 1 // 车辆已被注册
	RegResultNoVehicle       byte = 2 // 数据库中无该车辆
	RegResultTerminalExists  byte = 3 // 终端已被注册
	RegResultNoVehicleRecord byte = 4 // 终端已被注册（数据库中无该车辆）
)

// 坐标缩放因子
const CoordScaleFactor = 1000000.0

// 单帧最大长度
const MaxFrameSize = 1 * 1024 * 1024

// 错误
var (
	ErrDataTooShort    = errors.New("message data too short")
	ErrInvalidChecksum = errors.New("invalid checksum")
	ErrInvalidDelimiter = errors.New("invalid message delimiter")
)

// ==================== 版本差异 ====================

// Version808 表示808协议版本
type Version808 string

const (
	Version808_2011 Version808 = "2011"
	Version808_2013 Version808 = "2013"
	Version808_2019 Version808 = "2019"
)

// VersionFeature 版本特性
type VersionFeature struct {
	HasProtocolVer   bool // 消息头中是否有协议版本号字节(2019有)
	HasChromaParam   bool // 摄像头拍摄是否有色度参数(2019有)
	BatchItemLenSize int  // 批量位置项长度字节数(2019=2, 2011/2013=2)
	AuthHasIMEI      bool // 鉴权消息是否有IMEI(2019有)
}

// GetVersionFeature 获取版本特性
func GetVersionFeature(ver Version808) VersionFeature {
	switch ver {
	case Version808_2011:
		return VersionFeature{
			HasProtocolVer:   false,
			HasChromaParam:   false,
			BatchItemLenSize: 2,
			AuthHasIMEI:      false,
		}
	case Version808_2013:
		return VersionFeature{
			HasProtocolVer:   false,
			HasChromaParam:   false,
			BatchItemLenSize: 2,
			AuthHasIMEI:      false,
		}
	default: // 2019
		return VersionFeature{
			HasProtocolVer:   true,
			HasChromaParam:   true,
			BatchItemLenSize: 2,
			AuthHasIMEI:      true,
		}
	}
}

// VersionMsgIDs 每个版本支持的消息ID列表
// 用于前端动态过滤消息ID
type MsgInfo struct {
	ID   uint16
	Name string
	Dir  string // "up"=终端上行, "down"=平台下行, "both"=双向
}

// AllMessages 所有版本共有的消息列表（2019全集）
var AllMessages = []MsgInfo{
	{0x0001, "终端通用应答", "up"},
	{0x8001, "平台通用应答", "down"},
	{0x0002, "终端心跳", "up"},
	{0x0003, "终端注销", "up"},
	{0x8003, "终端注销应答", "down"},
	{0x0100, "终端注册", "up"},
	{0x8100, "终端注册应答", "down"},
	{0x0102, "终端鉴权", "up"},
	{0x8103, "设置参数", "down"},
	{0x0103, "设置参数应答", "up"},
	{0x8104, "查询参数", "down"},
	{0x0104, "查询参数应答", "up"},
	{0x8105, "终端控制", "down"},
	{0x8106, "查询指定参数", "down"},
	{0x8107, "终端属性查询", "down"},
	{0x0107, "终端属性应答", "up"},
	{0x8108, "终端升级", "down"},
	{0x0108, "终端升级应答", "up"},
	{0x0200, "位置上报", "up"},
	{0x8201, "位置查询", "down"},
	{0x0201, "位置查询应答", "up"},
	{0x8202, "临时位置跟踪", "down"},
	{0x0202, "临时位置跟踪应答", "up"},
	{0x8203, "人工确认报警", "down"},
	{0x8204, "电话簿设置", "down"},
	{0x8300, "文本下发", "down"},
	{0x8301, "事件设置", "down"},
	{0x0301, "事件应答", "up"},
	{0x8302, "提问下发/事件删除", "down"},
	{0x0302, "提问应答", "up"},
	{0x8303, "信息点下发", "down"},
	{0x8304, "信息服务", "down"},
	{0x8400, "超速报警设置", "down"},
	{0x0400, "超速报警", "up"},
	{0x8401, "疲劳驾驶设置", "down"},
	{0x0401, "疲劳驾驶报警", "up"},
	{0x8402, "区域路线报警设置", "down"},
	{0x8403, "区域路线报警删除", "down"},
	{0x8404, "电子运单设置", "down"},
	{0x8405, "电子运单删除", "down"},
	{0x8406, "电子运单上报", "down"},
	{0x8407, "电子运单应答", "up"},
	{0x8500, "车辆控制", "down"},
	{0x8600, "圆形区域设置", "down"},
	{0x8601, "圆形区域删除", "down"},
	{0x8602, "矩形区域设置", "down"},
	{0x8603, "矩形区域删除", "down"},
	{0x8604, "多边形区域设置", "down"},
	{0x8605, "多边形区域删除", "down"},
	{0x8606, "路线设置", "down"},
	{0x8607, "路线删除", "down"},
	{0x8608, "火源区域设置", "down"},
	{0x8609, "火源区域删除", "down"},
	{0x0500, "区域路线报警", "up"},
	{0x8700, "信息菜单设置", "down"},
	{0x0700, "信息菜单应答", "up"},
	{0x8701, "信息点推送", "down"},
	{0x8702, "电话回拨", "down"},
	{0x0702, "驾驶员身份", "up"},
	{0x8703, "短信转发", "down"},
	{0x0703, "短信转发应答", "up"},
	{0x0701, "电子运单", "up"},
	{0x0704, "批量位置上报", "up"},
	{0x0705, "CAN数据", "up"},
	{0x0801, "多媒体事件", "up"},
	{0x0802, "多媒体数据上传", "up"},
	{0x0803, "存储多媒体检索", "up"},
	{0x0804, "存储多媒体上传", "up"},
	{0x0805, "摄像头拍摄应答", "up"},
	{0x8801, "摄像头立即拍摄", "down"},
	{0x8802, "多媒体上传命令", "down"},
	{0x8803, "文件上传命令", "down"},
	{0x8804, "录音命令", "down"},
	{0x0900, "报警", "up"},
	{0x8900, "报警确认", "down"},
	{0x0901, "报警附件", "up"},
	{0x9001, "报警附件应答", "down"},
	{0x0A00, "终端RSA公钥", "up"},
	{0x8A00, "平台RSA公钥下发", "down"},
	{0x0B00, "计价器操作", "both"},
}

// MessagesByVersion 按版本过滤消息列表
// 2011版：不包含 0x8204, 0x8303, 0x8304, 0x8402-0x8407, 0x8608-0x8609, 0x0705, 0x0901, 0x9001, 0x0A00, 0x8A00, 0x0B00
// 2013版：在2011基础上增加 0x8402-0x8403, 0x0705, 0x0901, 0x9001
func MessagesByVersion(ver Version808) []MsgInfo {
	switch ver {
	case Version808_2011:
		exclude := map[uint16]bool{
			0x8204: true, 0x8303: true, 0x8304: true,
			0x8402: true, 0x8403: true, 0x8404: true, 0x8405: true, 0x8406: true, 0x8407: true,
			0x8608: true, 0x8609: true,
			0x0705: true, 0x0901: true, 0x9001: true,
			0x0A00: true, 0x8A00: true, 0x0B00: true,
		}
		var result []MsgInfo
		for _, m := range AllMessages {
			if !exclude[m.ID] {
				result = append(result, m)
			}
		}
		return result
	case Version808_2013:
		exclude := map[uint16]bool{
			0x8204: true, 0x8303: true, 0x8304: true,
			0x8404: true, 0x8405: true, 0x8406: true, 0x8407: true,
			0x8608: true, 0x8609: true,
			0x0A00: true, 0x8A00: true, 0x0B00: true,
		}
		var result []MsgInfo
		for _, m := range AllMessages {
			if !exclude[m.ID] {
				result = append(result, m)
			}
		}
		return result
	default:
		return AllMessages
	}
}
