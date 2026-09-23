package jt809

// ==================== JT/T 809-2019 完整消息ID常量 ====================
// 主链路（下级平台→上级平台）= UP
// 从链路（上级平台→下级平台）= DOWN

const (
	// ==================== 链路管理类 ====================
	MsgIDUpConnectReq    uint16 = 0x1001 // 主链路连接请求
	MsgIDUpConnectRsp    uint16 = 0x1002 // 主链路连接应答
	MsgIDUpDisconnectReq uint16 = 0x1003 // 主链路断开请求
	MsgIDUpDisconnectRsp uint16 = 0x1004 // 主链路断开应答
	MsgIDUpLinktestReq   uint16 = 0x1005 // 主链路连接保持请求
	MsgIDUpLinktestRsp   uint16 = 0x1006 // 主链路连接保持应答
	MsgIDDnDisconnectReq uint16 = 0x9003 // 从链路断开通知
	MsgIDDnDisconnectRsp uint16 = 0x9004 // 从链路断开应答
	MsgIDDnLinktestReq   uint16 = 0x9005 // 从链路连接保持请求
	MsgIDDnLinktestRsp   uint16 = 0x9006 // 从链路连接保持应答

	// ==================== 车辆报文类（主链路） ====================
	MsgIDUploadCarMsg      uint16 = 0x1200 // 车辆信息上报（主链路）
	MsgIDStartupMsg        uint16 = 0x1202 // 车辆启动（主链路）
	MsgIDShutdownMsg       uint16 = 0x1203 // 车辆关闭（主链路）
	MsgIDLocationMsg       uint16 = 0x1205 // 车辆定位上报（主链路）
	MsgIDAlarmMsg          uint16 = 0x1300 // 报警上报（主链路）
	MsgIDHistoryLocation   uint16 = 0x1206 // 历史定位上报

	// ==================== 车辆报文类（从链路） ====================
	MsgIDDownloadCarMsg    uint16 = 0x9200 // 车辆信息下发（从链路）
	MsgIDStartupAckMsg     uint16 = 0x9202 // 车辆启动应答（从链路）
	MsgIDShutdownAckMsg    uint16 = 0x9203 // 车辆关闭应答（从链路）
	MsgIDLocationAckMsg    uint16 = 0x9205 // 定位上报应答（从链路）
	MsgIDAlarmAckMsg       uint16 = 0x9300 // 报警应答（从链路）
	MsgIDHistoryLocationAck uint16 = 0x9206 // 历史定位应答

	// ==================== 车辆管理类（主链路） ====================
	MsgIDSyncVehicleMsg    uint16 = 0x1601 // 同步车辆请求（主链路）
	MsgIDSyncVehicleRsp    uint16 = 0x1602 // 同步车辆应答（主链路）
	MsgIDSubscribeMsg      uint16 = 0x1603 // 订阅车辆（主链路）
	MsgIDUnsubscribeMsg    uint16 = 0x1604 // 取消订阅（主链路）
	MsgIDQueryVehicleMsg   uint16 = 0x1605 // 查询车辆（主链路）
	MsgIDQueryVehicleRsp   uint16 = 0x1606 // 查询车辆应答（主链路）
	MsgIDBatchSubscribeMsg uint16 = 0x1607 // 批量订阅车辆
	MsgIDBatchUnsubscribeMsg uint16 = 0x1608 // 批量取消订阅

	// ==================== 车辆管理类（从链路） ====================
	MsgIDSyncVehicleDnMsg  uint16 = 0x9601 // 同步车辆请求（从链路）
	MsgIDSyncVehicleDnRsp  uint16 = 0x9602 // 同步车辆应答（从链路）
	MsgIDSubscribeDnMsg    uint16 = 0x9603 // 订阅车辆（从链路）
	MsgIDUnsubscribeDnMsg  uint16 = 0x9604 // 取消订阅（从链路）
	MsgIDQueryVehicleDnMsg uint16 = 0x9605 // 查询车辆（从链路）
	MsgIDQueryVehicleDnRsp uint16 = 0x9606 // 查询车辆应答（从链路）

	// ==================== 报警类（主链路） ====================
	MsgIDAlarmUpdateMsg    uint16 = 0x1301 // 报警更新上报
	MsgIDAlarmForwardMsg   uint16 = 0x1302 // 报警转发上报
	MsgIDAlarmStatistics   uint16 = 0x1303 // 报警统计上报

	// ==================== 驾驶员类（主链路） ====================
	MsgIDDriverInfoMsg     uint16 = 0x1701 // 驾驶员信息上报
	MsgIDDriverInfoRsp     uint16 = 0x1702 // 驾驶员信息应答

	// ==================== 电子运单类（主链路） ====================
	MsgIDEWaybillUpMsg     uint16 = 0x1801 // 电子运单上报
	MsgIDEWaybillDnMsg     uint16 = 0x9801 // 电子运单下发

	// ==================== 多媒体类（主链路） ====================
	MsgIDMultimediaSearchMsg  uint16 = 0x1401 // 多媒体检索请求
	MsgIDMultimediaSearchRsp  uint16 = 0x1402 // 多媒体检索应答
	MsgIDMultimediaUploadMsg  uint16 = 0x1403 // 多媒体上传请求
	MsgIDMultimediaUploadRsp  uint16 = 0x1404 // 多媒体上传应答
	MsgIDMultimediaDownloadMsg uint16 = 0x1405 // 多媒体下载请求
	MsgIDMultimediaDownloadRsp uint16 = 0x1406 // 多媒体下载应答
	MsgIDMultimediaForwardMsg uint16 = 0x1407 // 多媒体转发

	// ==================== 报文统计类（主链路） ====================
	MsgIDStatisticsMsg     uint16 = 0x1501 // 报文统计上报
	MsgIDStatisticsRsp     uint16 = 0x1502 // 报文统计应答

	// ==================== 车辆状态类（主链路） ====================
	MsgIDVehicleStateMsg   uint16 = 0x1201 // 车辆状态上报
	MsgIDVehicleStateAck   uint16 = 0x9201 // 车辆状态应答

	// ==================== 从链路下行控制 ====================
	MsgIDDownloadCmdMsg    uint16 = 0x9301 // 下发报文请求
	MsgIDDownloadCmdRsp    uint16 = 0x1304 // 下发报文应答
)

// ==================== 809版本定义 ====================
type Version809 string

const (
	Version809_2011 Version809 = "2011"
	Version809_2019 Version809 = "2019"
)

// MsgInfo809 809消息信息
type MsgInfo809 struct {
	ID   uint16
	Name string
	Dir  string // "up"=主链路, "down"=从链路
}

// AllMessages809 809协议全部消息列表
var AllMessages809 = []MsgInfo809{
	{0x1001, "主链路连接请求", "up"},
	{0x1002, "主链路连接应答", "up"},
	{0x1003, "主链路断开请求", "up"},
	{0x1004, "主链路断开应答", "up"},
	{0x1005, "主链路连接保持请求", "up"},
	{0x1006, "主链路连接保持应答", "up"},
	{0x9003, "从链路断开通知", "down"},
	{0x9004, "从链路断开应答", "down"},
	{0x9005, "从链路连接保持请求", "down"},
	{0x9006, "从链路连接保持应答", "down"},
	{0x1200, "车辆信息上报", "up"},
	{0x1201, "车辆状态上报", "up"},
	{0x1202, "车辆启动", "up"},
	{0x1203, "车辆关闭", "up"},
	{0x1205, "车辆定位上报", "up"},
	{0x1206, "历史定位上报", "up"},
	{0x1300, "报警上报", "up"},
	{0x1301, "报警更新上报", "up"},
	{0x1302, "报警转发上报", "up"},
	{0x1303, "报警统计上报", "up"},
	{0x1304, "下发报文应答", "up"},
	{0x9200, "车辆信息下发", "down"},
	{0x9201, "车辆状态应答", "down"},
	{0x9202, "车辆启动应答", "down"},
	{0x9203, "车辆关闭应答", "down"},
	{0x9205, "定位上报应答", "down"},
	{0x9206, "历史定位应答", "down"},
	{0x9300, "报警应答", "down"},
	{0x9301, "下发报文请求", "down"},
	{0x1601, "同步车辆请求", "up"},
	{0x1602, "同步车辆应答", "up"},
	{0x1603, "订阅车辆", "up"},
	{0x1604, "取消订阅", "up"},
	{0x1605, "查询车辆", "up"},
	{0x1606, "查询车辆应答", "up"},
	{0x1607, "批量订阅车辆", "up"},
	{0x1608, "批量取消订阅", "up"},
	{0x9601, "同步车辆请求(从链路)", "down"},
	{0x9602, "同步车辆应答(从链路)", "down"},
	{0x9603, "订阅车辆(从链路)", "down"},
	{0x9604, "取消订阅(从链路)", "down"},
	{0x9605, "查询车辆(从链路)", "down"},
	{0x9606, "查询车辆应答(从链路)", "down"},
	{0x1701, "驾驶员信息上报", "up"},
	{0x1702, "驾驶员信息应答", "up"},
	{0x1801, "电子运单上报", "up"},
	{0x9801, "电子运单下发", "down"},
	{0x1401, "多媒体检索请求", "up"},
	{0x1402, "多媒体检索应答", "up"},
	{0x1403, "多媒体上传请求", "up"},
	{0x1404, "多媒体上传应答", "up"},
	{0x1405, "多媒体下载请求", "up"},
	{0x1406, "多媒体下载应答", "up"},
	{0x1407, "多媒体转发", "up"},
	{0x1501, "报文统计上报", "up"},
	{0x1502, "报文统计应答", "up"},
}

// MessagesByVersion809 按版本过滤809消息
func MessagesByVersion809(ver Version809) []MsgInfo809 {
	switch ver {
	case Version809_2011:
		// 2011版不包含: 驾驶员信息、电子运单、多媒体检索、报文统计、批量订阅
		exclude := map[uint16]bool{
			0x1701: true, 0x1702: true,
			0x1801: true, 0x9801: true,
			0x1401: true, 0x1402: true, 0x1403: true, 0x1404: true,
			0x1405: true, 0x1406: true, 0x1407: true,
			0x1501: true, 0x1502: true,
			0x1607: true, 0x1608: true,
			0x1201: true, 0x9201: true,
			0x1206: true, 0x9206: true,
			0x1301: true, 0x1302: true, 0x1303: true,
			0x9004: true, 0x9006: true,
			0x9601: true, 0x9602: true, 0x9603: true, 0x9604: true, 0x9605: true, 0x9606: true,
			0x9301: true, 0x1304: true,
		}
		var result []MsgInfo809
		for _, m := range AllMessages809 {
			if !exclude[m.ID] {
				result = append(result, m)
			}
		}
		return result
	default:
		return AllMessages809
	}
}
