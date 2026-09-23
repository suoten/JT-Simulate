package engine

import (
	"fmt"

	"github.com/suoten/jt-simulate/pkg/codec/jt808"
)

// KnowledgeBase 协议知识库
type KnowledgeBase struct {
	protocols map[string]*ProtocolInfo
}

// ProtocolInfo 协议信息
type ProtocolInfo struct {
	Name        string              `json:"name"`
	FullName    string              `json:"full_name"`
	Version     string              `json:"version"`
	Description string              `json:"description"`
	Messages    []MessageInfo       `json:"messages"`
	Features    []string            `json:"features"`
}

// MessageInfo 消息信息
type MessageInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Direction   string `json:"direction"` // up, down, both
	Description string `json:"description"`
}

// NewKnowledgeBase 创建知识库
func NewKnowledgeBase() *KnowledgeBase {
	kb := &KnowledgeBase{
		protocols: make(map[string]*ProtocolInfo),
	}
	kb.initJT808()
	kb.initJT809()
	kb.initJT1078()
	kb.initJT905()
	kb.initJT1045()
	kb.initJT1253()
	kb.initGBT32960()
	return kb
}

// GetProtocol 获取协议信息
func (kb *KnowledgeBase) GetProtocol(name string) (*ProtocolInfo, error) {
	info, ok := kb.protocols[name]
	if !ok {
		return nil, fmt.Errorf("protocol %s not found", name)
	}
	return info, nil
}

// ListProtocols 列出所有协议
func (kb *KnowledgeBase) ListProtocols() []*ProtocolInfo {
	list := make([]*ProtocolInfo, 0, len(kb.protocols))
	for _, p := range kb.protocols {
		list = append(list, p)
	}
	return list
}

func (kb *KnowledgeBase) initJT808() {
	kb.protocols["jt808"] = &ProtocolInfo{
		Name:    "jt808",
		FullName: "JT/T 808-2019",
		Version: "2019",
		Description: "道路运输车辆卫星定位系统终端通讯协议及数据格式",
		Features: []string{"终端注册", "终端鉴权", "位置上报", "报警", "区域设置", "文本下发", "电话回拨", "多媒体"},
		Messages: []MessageInfo{
			{"0x0001", "终端通用应答", "up", "终端对平台消息的通用应答"},
			{"0x8001", "平台通用应答", "down", "平台对终端消息的通用应答"},
			{"0x0002", "终端心跳", "up", "终端定时发送的心跳消息"},
			{"0x0100", "终端注册", "up", "终端向平台注册"},
			{"0x8100", "终端注册应答", "down", "平台对注册的应答"},
			{"0x0102", "终端鉴权", "up", "终端向平台鉴权"},
			{"0x0200", "位置上报", "up", "终端上报位置信息"},
			{"0x0704", "批量位置上报", "up", "终端批量上报历史位置"},
			{"0x0702", "驾驶员身份", "up", "上报驾驶员身份信息"},
			{"0x0900", "报警", "up", "终端上报报警"},
			{"0x8103", "设置参数", "down", "平台设置终端参数"},
			{"0x8105", "终端控制", "down", "平台控制终端"},
			{"0x8201", "查询位置", "down", "平台查询终端位置"},
			{"0x8300", "文本下发", "down", "平台下发文本信息"},
			{"0x8500", "车辆控制", "down", "平台控制车辆（断油断电等）"},
			{"0x8600", "圆形区域设置", "down", "设置圆形电子围栏"},
			{"0x8602", "矩形区域设置", "down", "设置矩形电子围栏"},
			{"0x0801", "多媒体事件", "up", "终端上报多媒体事件"},
			{"0x8801", "摄像头立即拍摄", "down", "平台命令终端立即拍摄"},
		},
	}
}

func (kb *KnowledgeBase) initJT809() {
	kb.protocols["jt809"] = &ProtocolInfo{
		Name:    "jt809",
		FullName: "JT/T 809-2019",
		Version: "2019",
		Description: "道路运输车辆卫星定位系统平台数据交换",
		Features: []string{"平台间数据交换", "车辆信息同步", "报警转发", "车辆订阅"},
		Messages: []MessageInfo{
			{"0x1001", "下级平台请求连接", "up", "下级平台向上级平台请求连接"},
			{"0x1002", "连接应答", "down", "上级平台对连接请求的应答"},
			{"0x1005", "链路保持", "up", "下级平台发送链路保持"},
			{"0x1200", "车辆信息上报", "up", "下级平台上报车辆信息"},
			{"0x1205", "上报车辆定位", "up", "上报车辆定位信息"},
			{"0x1300", "上报报警", "up", "上报报警信息"},
			{"0x1603", "订阅车辆", "down", "上级平台订阅车辆"},
			{"0x1604", "取消订阅", "down", "取消订阅车辆"},
		},
	}
}

func (kb *KnowledgeBase) initJT1078() {
	kb.protocols["jt1078"] = &ProtocolInfo{
		Name:    "jt1078",
		FullName: "JT/T 1078-2016",
		Version: "2016",
		Description: "道路运输车辆卫星定位系统视频通信协议",
		Features: []string{"实时音视频", "历史回放", "云台控制", "音视频下载", "RTP传输"},
		Messages: []MessageInfo{
			{"0x9101", "实时音视频请求", "down", "平台请求终端实时音视频"},
			{"0x9102", "实时音视频控制", "down", "控制实时音视频（暂停/恢复等）"},
			{"0x9201", "音视频回放请求", "down", "请求历史音视频回放"},
			{"0x9203", "回放控制", "down", "控制回放（快进/快退/拖拽）"},
			{"0x9301", "云台控制", "down", "控制云台方向和焦距"},
			{"0x9501", "音视频参数设置", "down", "设置音视频编码参数"},
		},
	}
}

func (kb *KnowledgeBase) initJT905() {
	kb.protocols["jt905"] = &ProtocolInfo{
		Name:    "jt905",
		FullName: "JT/T 905-2014",
		Version: "2014",
		Description: "出租汽车服务管理信息系统数据交换与共享",
		Features: []string{"计价器状态", "运营数据", "调度信息", "广告信息"},
		Messages: []MessageInfo{
			{"0x0B00", "计价器状态", "up", "上报计价器状态"},
			{"0x0B01", "运营数据上报", "up", "上报上下车运营数据"},
			{"0x8B02", "调度信息", "down", "平台下发调度信息"},
		},
	}
}

func (kb *KnowledgeBase) initJT1045() {
	kb.protocols["jt1045"] = &ProtocolInfo{
		Name:    "jt1045",
		FullName: "JT/T 1045-2016",
		Version: "2016",
		Description: "道路运输车辆卫星定位系统主动安全智能防控系统",
		Features: []string{"ADAS报警", "DSM驾驶员状态", "BSD盲区检测", "胎压监测"},
		Messages: []MessageInfo{
			{"0x0901", "ADAS报警", "up", "高级驾驶辅助报警"},
			{"0x1205", "DSM报警", "up", "驾驶员状态监控报警"},
			{"0x1209", "盲区报警", "up", "BSD盲区检测报警"},
		},
	}
}

func (kb *KnowledgeBase) initJT1253() {
	kb.protocols["jt1253"] = &ProtocolInfo{
		Name:    "jt1253",
		FullName: "JT/T 1253-2019",
		Version: "2019",
		Description: "危险货物道路运输车辆卫星定位系统平台技术要求",
		Features: []string{"电子运单", "危险品报警", "卸载报告"},
		Messages: []MessageInfo{
			{"0x0D01", "电子运单上报", "up", "上报电子运单信息"},
			{"0x0D02", "危险品报警", "up", "危险品运输报警"},
		},
	}
}

func (kb *KnowledgeBase) initGBT32960() {
	kb.protocols["gbt32960"] = &ProtocolInfo{
		Name:    "gbt32960",
		FullName: "GB/T 32960-2016",
		Version: "2016",
		Description: "电动汽车远程服务与管理系统技术规范",
		Features: []string{"车辆登入", "实时数据上报", "充电数据", "故障报警"},
		Messages: []MessageInfo{
			{"0x01", "车辆登入", "up", "新能源汽车登入"},
			{"0x02", "实时信息上报", "up", "实时上报车辆数据"},
			{"0x03", "车辆登出", "up", "车辆登出"},
			{"0x07", "心跳", "up", "心跳消息"},
		},
	}
}

// SequenceDiagram 时序图
type SequenceDiagram struct {
	Title    string          `json:"title"`
	Protocol string          `json:"protocol"`
	Actors   []string        `json:"actors"` // "终端", "平台"
	Steps    []DiagramStep   `json:"steps"`
}

// DiagramStep 时序图步骤
type DiagramStep struct {
	From      string `json:"from"`
	To        string `json:"to"`
	MsgID     string `json:"msg_id"`
	MsgName   string `json:"msg_name"`
	Direction string `json:"direction"` // ->, <-, <->, -- (note)
	Note      string `json:"note,omitempty"`
}

// GetSequenceDiagram 获取时序图
func (kb *KnowledgeBase) GetSequenceDiagram(scenario string) (*SequenceDiagram, error) {
	switch scenario {
	case "register":
		return &SequenceDiagram{
			Title: "终端注册上线流程",
			Protocol: "jt808",
			Actors: []string{"终端", "平台"},
			Steps: []DiagramStep{
				{"终端", "平台", "0x0100", "终端注册", "->", "包含省域ID、制造商、终端型号等"},
				{"平台", "终端", "0x8100", "注册应答", "<-", "返回鉴权码"},
				{"终端", "平台", "0x0102", "终端鉴权", "->", "使用注册应答中的鉴权码"},
				{"平台", "终端", "0x8001", "通用应答", "<-", "鉴权成功"},
				{"终端", "平台", "0x0002", "心跳", "->", "定时发送保持连接"},
				{"终端", "平台", "0x0200", "位置上报", "->", "定时上报位置信息"},
			},
		}, nil
	case "alarm":
		return &SequenceDiagram{
			Title: "报警处理流程",
			Protocol: "jt808",
			Actors: []string{"终端", "平台"},
			Steps: []DiagramStep{
				{"终端", "平台", "0x0200", "位置上报(含报警标志)", "->", "报警标志非0表示有报警"},
				{"平台", "终端", "0x8001", "通用应答", "<-", "确认收到"},
				{"平台", "平台", "", "触发报警处理", "--", "根据报警类型处理"},
				{"平台", "终端", "0x8203", "人工确认报警", "<-", "确认报警"},
				{"终端", "平台", "0x0001", "通用应答", "->", "终端确认"},
			},
		}, nil
	case "video":
		return &SequenceDiagram{
			Title: "实时视频请求流程",
			Protocol: "jt1078",
			Actors: []string{"终端", "平台"},
			Steps: []DiagramStep{
				{"平台", "终端", "0x9101", "实时音视频请求", "->", "指定通道和数据类型"},
				{"终端", "平台", "0x1200", "RTP数据", "->", "开始传输音视频RTP包"},
				{"平台", "终端", "0x9102", "音视频控制", "<-", "暂停/恢复/停止"},
				{"终端", "平台", "0x1200", "RTP数据", "->", "持续传输"},
				{"平台", "终端", "0x9102", "停止传输", "<-", "停止音视频"},
			},
		}, nil
	default:
		return nil, fmt.Errorf("unknown scenario: %s", scenario)
	}
}

// unused guard
var _ = jt808.MsgIDHeartbeat
