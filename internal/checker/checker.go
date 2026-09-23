package checker

import (
	"fmt"
	"strings"

	"github.com/suoten/jt-simulate/pkg/codec/jt808"
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
