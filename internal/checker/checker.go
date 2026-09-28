package checker

import (
	"fmt"
	"strings"

	gbt32960 "github.com/suoten/jt-simulate/pkg/codec/gbt32960"
	"github.com/suoten/jt-simulate/pkg/codec/jt808"
	"github.com/suoten/jt-simulate/pkg/codec/jt809"
	"github.com/suoten/jt-simulate/pkg/types"
)

// CheckResult 检查结果
type CheckResult struct {
	Total     int           `json:"total"`
	Passed    int           `json:"passed"`
	Failed    int           `json:"failed"`
	Warnings  int           `json:"warnings"`
	Score     int           `json:"score"` // 0-100
	Grade     string        `json:"grade"` // A/B/C/D/F
	Items     []CheckItem   `json:"items"`
}

// CheckItem 检查项
type CheckItem struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
	Level    string `json:"level"` // error, warning, info
	Passed   bool   `json:"passed"`
	Message  string `json:"message"`
}

// Checker 合规检查器
type Checker struct{}

// New 创建检查器
func New() *Checker {
	return &Checker{}
}

// CheckMessage 检查单条消息
func (c *Checker) CheckMessage(msg *types.Message) []CheckItem {
	var items []CheckItem

	switch msg.Header.MsgID {
	case jt808.MsgIDLocation:
		items = c.checkLocation(msg)
	case jt808.MsgIDRegister:
		items = c.checkRegister(msg)
	case jt808.MsgIDAuth:
		items = c.checkAuth(msg)
	case jt808.MsgIDHeartbeat:
		items = c.checkHeartbeat(msg)
	default:
		items = c.checkGeneric(msg)
	}

	return items
}

// CheckRaw 检查原始 hex 报文（支持多协议）
func (c *Checker) CheckRaw(protocol, hexStr string) (*CheckResult, error) {
	hexStr = strings.TrimSpace(hexStr)
	hexStr = strings.ReplaceAll(hexStr, " ", "")

	switch protocol {
	case "jt808", "jt1078", "jt905", "jt1045", "jt1253":
		return c.checkRawJT808(hexStr, protocol)
	case "jt809":
		return c.checkRawJT809(hexStr)
	case "gbt32960":
		return c.checkRawGBT32960(hexStr)
	default:
		return c.checkRawJT808(hexStr, "jt808")
	}
}

// checkRawJT808 检查 JT808 帧格式报文
func (c *Checker) checkRawJT808(hexStr, protocol string) (*CheckResult, error) {
	data, err := hexDecode(hexStr)
	if err != nil {
		return &CheckResult{
			Items: []CheckItem{{ID: "FMT-001", Name: "Hex解码", Category: "格式", Level: "error", Passed: false, Message: err.Error()}},
		}, nil
	}

	codec := jt808.NewCodec()
	msg, err := codec.Decode(data)
	if err != nil {
		return &CheckResult{
			Items: []CheckItem{{ID: "DEC-001", Name: "报文解码", Category: "解码", Level: "error", Passed: false, Message: "解码失败: " + err.Error()}},
		}, nil
	}

	items := c.CheckMessage(msg)

	// 额外的帧级检查
	items = append(items, c.checkFrame(data)...)

	return c.GenerateReport(items), nil
}

// checkRawJT809 检查 JT809 帧格式报文
func (c *Checker) checkRawJT809(hexStr string) (*CheckResult, error) {
	data, err := hexDecode(hexStr)
	if err != nil {
		return &CheckResult{
			Items: []CheckItem{{ID: "FMT-001", Name: "Hex解码", Category: "格式", Level: "error", Passed: false, Message: err.Error()}},
		}, nil
	}

	codec := jt809.NewCodec()
	h, bodyData, err := codec.Decode(data)
	if err != nil {
		return &CheckResult{
			Items: []CheckItem{{ID: "DEC-001", Name: "报文解码", Category: "解码", Level: "error", Passed: false, Message: "解码失败: " + err.Error()}},
		}, nil
	}

	var items []CheckItem

	// 检查消息序列号
	snOK := h.MsgSN > 0
	items = append(items, CheckItem{
		ID: "J809-001", Name: "消息序号", Category: "JT809",
		Level: "warning", Passed: snOK,
		Message: fmt.Sprintf("消息序号: %d", h.MsgSN),
	})

	// 检查 GNSSCenterID
	centerOK := h.GNSSCenterID > 0
	items = append(items, CheckItem{
		ID: "J809-002", Name: "GNSS中心ID", Category: "JT809",
		Level: "warning", Passed: centerOK,
		Message: fmt.Sprintf("GNSS中心ID: %d", h.GNSSCenterID),
	})

	// 检查消息体长度
	bodyOK := len(bodyData) > 0 || h.MsgID == jt809.MsgIDUpLinktestReq
	items = append(items, CheckItem{
		ID: "J809-003", Name: "消息体", Category: "JT809",
		Level: "info", Passed: bodyOK,
		Message: fmt.Sprintf("消息体长度: %d bytes", len(bodyData)),
	})

	return c.GenerateReport(items), nil
}

// checkRawGBT32960 检查 GB/T 32960 帧格式报文
func (c *Checker) checkRawGBT32960(hexStr string) (*CheckResult, error) {
	data, err := hexDecode(hexStr)
	if err != nil {
		return &CheckResult{
			Items: []CheckItem{{ID: "FMT-001", Name: "Hex解码", Category: "格式", Level: "error", Passed: false, Message: err.Error()}},
		}, nil
	}

	pkt, err := gbt32960.Decode(data)
	if err != nil {
		return &CheckResult{
			Items: []CheckItem{{ID: "DEC-001", Name: "报文解码", Category: "解码", Level: "error", Passed: false, Message: "解码失败: " + err.Error()}},
		}, nil
	}

	var items []CheckItem

	// 检查 VIN
	vinOK := len(pkt.Header.VIN) == 17
	items = append(items, CheckItem{
		ID: "GB-001", Name: "VIN长度", Category: "GB32960",
		Level: "error", Passed: vinOK,
		Message: fmt.Sprintf("VIN: '%s' (17位)", pkt.Header.VIN),
	})

	// 检查命令标识
	cmdOK := pkt.Header.CmdFlag >= 0x01 && pkt.Header.CmdFlag <= 0x07
	items = append(items, CheckItem{
		ID: "GB-002", Name: "命令标识", Category: "GB32960",
		Level: "error", Passed: cmdOK,
		Message: fmt.Sprintf("命令标识: 0x%02X (%s)", pkt.Header.CmdFlag, gbt32960.MsgTypeString(pkt.Header.CmdFlag)),
	})

	// 检查加密方式
	encOK := pkt.Header.EncryptType <= 3
	items = append(items, CheckItem{
		ID: "GB-003", Name: "加密方式", Category: "GB32960",
		Level: "warning", Passed: encOK,
		Message: fmt.Sprintf("加密方式: %d", pkt.Header.EncryptType),
	})

	return c.GenerateReport(items), nil
}

// checkFrame 帧级检查
func (c *Checker) checkFrame(data []byte) []CheckItem {
	var items []CheckItem

	// 检查首尾分隔符
	delimOK := len(data) >= 2 && data[0] == 0x7E && data[len(data)-1] == 0x7E
	items = append(items, CheckItem{
		ID: "FRM-001", Name: "帧分隔符", Category: "帧结构",
		Level: "error", Passed: delimOK,
		Message: "首尾必须为0x7E",
	})

	// 检查最小帧长
	lenOK := len(data) >= 15 // 最小帧长
	items = append(items, CheckItem{
		ID: "FRM-002", Name: "帧长度", Category: "帧结构",
		Level: "error", Passed: lenOK,
		Message: fmt.Sprintf("帧长: %d bytes", len(data)),
	})

	// 校验码验证
	codec := jt808.NewCodec()
	checksumOK := codec.VerifyChecksum(data[1 : len(data)-1])
	items = append(items, CheckItem{
		ID: "FRM-003", Name: "校验码", Category: "帧结构",
		Level: "error", Passed: checksumOK,
		Message: fmt.Sprintf("XOR校验: %v", checksumOK),
	})

	return items
}

// checkLocation 检查位置上报
func (c *Checker) checkLocation(msg *types.Message) []CheckItem {
	var items []CheckItem
	loc, ok := msg.Body.(*jt808.LocationMessage)
	if !ok {
		items = append(items, CheckItem{
			ID: "LOC-001", Name: "消息体类型", Category: "位置上报",
			Level: "error", Passed: false, Message: "消息体不是LocationMessage类型",
		})
		return items
	}

	// 检查纬度范围
	latOK := loc.Latitude >= -90 && loc.Latitude <= 90
	items = append(items, CheckItem{
		ID: "LOC-010", Name: "纬度范围", Category: "位置上报",
		Level: "error", Passed: latOK,
		Message: fmt.Sprintf("纬度 %.6f, 范围应 [-90, 90]", loc.Latitude),
	})

	// 检查经度范围
	lonOK := loc.Longitude >= -180 && loc.Longitude <= 180
	items = append(items, CheckItem{
		ID: "LOC-011", Name: "经度范围", Category: "位置上报",
		Level: "error", Passed: lonOK,
		Message: fmt.Sprintf("经度 %.6f, 范围应 [-180, 180]", loc.Longitude),
	})

	// 检查速度范围 (0.1km/h, max 300km/h = 3000)
	speedOK := loc.Speed <= 3000
	items = append(items, CheckItem{
		ID: "LOC-012", Name: "速度范围", Category: "位置上报",
		Level: "warning", Passed: speedOK,
		Message: fmt.Sprintf("速度 %.1f km/h", float64(loc.Speed)/10),
	})

	// 检查方向范围 (0-359)
	dirOK := loc.Direction <= 359
	items = append(items, CheckItem{
		ID: "LOC-013", Name: "方向范围", Category: "位置上报",
		Level: "error", Passed: dirOK,
		Message: fmt.Sprintf("方向 %d, 范围应 [0, 359]", loc.Direction),
	})

	// 检查时间格式 (YYMMDDHHmmss, 12位)
	timeOK := len(loc.Time) == 12
	items = append(items, CheckItem{
		ID: "LOC-014", Name: "时间格式", Category: "位置上报",
		Level: "error", Passed: timeOK,
		Message: fmt.Sprintf("时间 '%s', 应为12位BCD (YYMMDDHHmmss)", loc.Time),
	})

	// 检查报警标志
	if loc.AlarmFlag != 0 {
		items = append(items, CheckItem{
			ID: "LOC-015", Name: "报警标志", Category: "位置上报",
			Level: "info", Passed: true,
			Message: fmt.Sprintf("报警标志: 0x%08X", loc.AlarmFlag),
		})
	}

	// 检查海拔范围
	altOK := loc.Altitude <= 30000 // 30km
	items = append(items, CheckItem{
		ID: "LOC-016", Name: "海拔范围", Category: "位置上报",
		Level: "warning", Passed: altOK,
		Message: fmt.Sprintf("海拔 %d m", loc.Altitude),
	})

	return items
}

// checkRegister 检查注册消息
func (c *Checker) checkRegister(msg *types.Message) []CheckItem {
	var items []CheckItem
	reg, ok := msg.Body.(*jt808.RegisterMessage)
	if !ok {
		items = append(items, CheckItem{
			ID: "REG-001", Name: "消息体类型", Category: "终端注册",
			Level: "error", Passed: false, Message: "消息体不是RegisterMessage类型",
		})
		return items
	}

	// 检查省域ID
	provOK := reg.ProvinceID >= 11 && reg.ProvinceID <= 99
	items = append(items, CheckItem{
		ID: "REG-010", Name: "省域ID", Category: "终端注册",
		Level: "warning", Passed: provOK,
		Message: fmt.Sprintf("省域ID: %d", reg.ProvinceID),
	})

	// 检查制造商
	manuOK := len(reg.Manufacturer) > 0 && len(reg.Manufacturer) <= 5
	items = append(items, CheckItem{
		ID: "REG-011", Name: "制造商ID", Category: "终端注册",
		Level: "error", Passed: manuOK,
		Message: fmt.Sprintf("制造商: '%s' (1-5字节)", reg.Manufacturer),
	})

	// 检查终端型号
	modelOK := len(reg.TerminalModel) > 0 && len(reg.TerminalModel) <= 20
	items = append(items, CheckItem{
		ID: "REG-012", Name: "终端型号", Category: "终端注册",
		Level: "error", Passed: modelOK,
		Message: fmt.Sprintf("终端型号: '%s' (1-20字节)", reg.TerminalModel),
	})

	// 检查终端ID
	tidOK := len(reg.TerminalID) > 0 && len(reg.TerminalID) <= 7
	items = append(items, CheckItem{
		ID: "REG-013", Name: "终端ID", Category: "终端注册",
		Level: "error", Passed: tidOK,
		Message: fmt.Sprintf("终端ID: '%s' (1-7字节)", reg.TerminalID),
	})

	// 检查车牌颜色
	colorOK := reg.PlateColor >= 0 && reg.PlateColor <= 5
	items = append(items, CheckItem{
		ID: "REG-014", Name: "车牌颜色", Category: "终端注册",
		Level: "warning", Passed: colorOK,
		Message: fmt.Sprintf("车牌颜色: %d (0=未上牌,1=蓝,2=黄,3=黑,4=白,5=其他)", reg.PlateColor),
	})

	return items
}

// checkAuth 检查鉴权消息
func (c *Checker) checkAuth(msg *types.Message) []CheckItem {
	var items []CheckItem
	auth, ok := msg.Body.(*jt808.AuthMessage)
	if !ok {
		items = append(items, CheckItem{
			ID: "AUTH-001", Name: "消息体类型", Category: "终端鉴权",
			Level: "error", Passed: false, Message: "消息体不是AuthMessage类型",
		})
		return items
	}

	authOK := len(auth.AuthCode) > 0
	items = append(items, CheckItem{
		ID: "AUTH-010", Name: "鉴权码", Category: "终端鉴权",
		Level: "error", Passed: authOK,
		Message: fmt.Sprintf("鉴权码长度: %d", len(auth.AuthCode)),
	})

	return items
}

// checkHeartbeat 检查心跳
func (c *Checker) checkHeartbeat(msg *types.Message) []CheckItem {
	return []CheckItem{
		{
			ID: "HB-001", Name: "心跳消息体", Category: "心跳",
			Level: "info", Passed: true, Message: "心跳消息体应为空",
		},
	}
}

// checkGeneric 通用检查
func (c *Checker) checkGeneric(msg *types.Message) []CheckItem {
	var items []CheckItem

	// 检查手机号
	phoneOK := len(msg.Header.Phone) == 12 && strings.HasPrefix(msg.Header.Phone, "0")
	items = append(items, CheckItem{
		ID: "GEN-001", Name: "终端手机号", Category: "通用",
		Level: "warning", Passed: phoneOK,
		Message: fmt.Sprintf("手机号: '%s' (应为12位, 以0开头)", msg.Header.Phone),
	})

	// 检查消息ID
	msgIDOK := msg.Header.MsgID != 0
	items = append(items, CheckItem{
		ID: "GEN-002", Name: "消息ID", Category: "通用",
		Level: "error", Passed: msgIDOK,
		Message: fmt.Sprintf("消息ID: 0x%04X (%s)", msg.Header.MsgID, jt808.MsgName(msg.Header.MsgID)),
	})

	return items
}

// GenerateReport 生成检查报告
func (c *Checker) GenerateReport(items []CheckItem) *CheckResult {
	result := &CheckResult{
		Total: len(items),
		Items: items,
	}

	for _, item := range items {
		if item.Passed {
			result.Passed++
		} else {
			if item.Level == "error" {
				result.Failed++
			} else {
				result.Warnings++
			}
		}
	}

	if result.Total > 0 {
		result.Score = result.Passed * 100 / result.Total
	} else {
		result.Score = 100
	}

	switch {
	case result.Score >= 90:
		result.Grade = "A"
	case result.Score >= 80:
		result.Grade = "B"
	case result.Score >= 70:
		result.Grade = "C"
	case result.Score >= 60:
		result.Grade = "D"
	default:
		result.Grade = "F"
	}

	return result
}

// hexDecode 简单 hex 解码
func hexDecode(s string) ([]byte, error) {
	// 移除空格和 0x 前缀
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "0x", "")
	s = strings.ReplaceAll(s, "0X", "")

	if len(s)%2 != 0 {
		return nil, fmt.Errorf("odd hex length: %d", len(s))
	}

	data := make([]byte, len(s)/2)
	for i := 0; i < len(data); i++ {
		var b byte
		_, err := fmt.Sscanf(s[i*2:i*2+2], "%x", &b)
		if err != nil {
			return nil, fmt.Errorf("hex decode error at position %d: %w", i, err)
		}
		data[i] = b
	}
	return data, nil
}
