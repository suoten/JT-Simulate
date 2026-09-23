package checker

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/suoten/jt-simulate/pkg/codec/jt808"
	"github.com/suoten/jt-simulate/pkg/types"
)

// FuzzCase Fuzz测试用例
type FuzzCase struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"` // boundary, malformed, injection, overflow
	Hex         string `json:"hex"`
}

// FuzzResult Fuzz测试结果
type FuzzResult struct {
	TotalCases   int           `json:"total_cases"`
	CrashCases   int           `json:"crash_cases"`
	ErrorCases   int           `json:"error_cases"`
	SuccessCases int           `json:"success_cases"`
	Cases        []FuzzCaseResult `json:"cases"`
}

// FuzzCaseResult 单个Fuzz测试结果
type FuzzCaseResult struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	Hex      string `json:"hex"`
	Result   string `json:"result"` // success, error, crash
	Error    string `json:"error,omitempty"`
}

// Fuzzer Fuzz测试器
type Fuzzer struct {
	codec *jt808.JT808Codec
	rng   *rand.Rand
}

// NewFuzzer 创建Fuzz测试器
func NewFuzzer() *Fuzzer {
	return &Fuzzer{
		codec: jt808.NewCodec(),
		rng:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// GenerateCases 生成Fuzz测试用例
func (f *Fuzzer) GenerateCases() []FuzzCase {
	var cases []FuzzCase

	// 1. 边界值测试
	cases = append(cases, FuzzCase{
		Name: "空帧", Category: "boundary",
		Description: "只包含分隔符的空帧", Hex: "7E7E",
	})
	cases = append(cases, FuzzCase{
		Name: "最小帧", Category: "boundary",
		Description: "最小可能的合法帧", Hex: "7E000000000000000000000000007E",
	})
	cases = append(cases, FuzzCase{
		Name: "超长帧", Category: "overflow",
		Description: "超过最大帧长的帧", Hex: "7E" + randomHex(2048) + "7E",
	})

	// 2. 畸形消息
	cases = append(cases, FuzzCase{
		Name: "无分隔符", Category: "malformed",
		Description: "没有0x7E分隔符", Hex: "0200801C01013800001234000100",
	})
	cases = append(cases, FuzzCase{
		Name: "错误校验码", Category: "malformed",
		Description: "校验码错误", Hex: "7E0200801C01013800001234000100000000000000000260F7B406F015580000025800B4260922110355FF7E",
	})
	cases = append(cases, FuzzCase{
		Name: "非法转义", Category: "malformed",
		Description: "尾部0x7D无转义字节", Hex: "7E0200801C0101380000123400017D7E",
	})
	cases = append(cases, FuzzCase{
		Name: "消息头过短", Category: "malformed",
		Description: "消息头不足12字节", Hex: "7E0200000538007E",
	})

	// 3. 注入测试
	cases = append(cases, FuzzCase{
		Name: "非法消息ID", Category: "injection",
		Description: "使用未定义的消息ID", Hex: "7EFFFF001C01013800001234000100000000000000000260F7B406F015580000025800B4260922110355XX7E",
	})
	cases = append(cases, FuzzCase{
		Name: "非法手机号", Category: "injection",
		Description: "手机号包含非BCD字符", Hex: "7E0200801C01FF00000000000100000000000000000260F7B406F015580000025800B4260922110355XX7E",
	})

	// 4. 随机突变
	for i := 0; i < 10; i++ {
		base := "7E0200801C01013800001234000100000000000000000260F7B406F015580000025800B4260922110355BF7E"
		mutated := mutateHex(base, f.rng)
		cases = append(cases, FuzzCase{
			Name: fmt.Sprintf("随机突变 #%d", i+1), Category: "mutation",
			Description: "对合法位置上报报文进行随机字节突变", Hex: mutated,
		})
	}

	return cases
}

// Run 执行Fuzz测试
func (f *Fuzzer) Run() *FuzzResult {
	cases := f.GenerateCases()
	result := &FuzzResult{TotalCases: len(cases)}

	for _, c := range cases {
		r := FuzzCaseResult{
			Name:     c.Name,
			Category: c.Category,
			Hex:      c.Hex,
		}

		// 尝试解码
		hexStr := c.Hex
		// 替换XX为随机字节
		for i := 0; i < len(hexStr); i++ {
			if hexStr[i] == 'X' {
				hexStr = hexStr[:i] + fmt.Sprintf("%x", f.rng.Intn(16)) + hexStr[i+1:]
			}
		}

		_, err := f.decodeHex(hexStr)
		if err == nil {
			r.Result = "success"
			result.SuccessCases++
		} else {
			r.Result = "error"
			r.Error = err.Error()
			result.ErrorCases++
		}

		result.Cases = append(result.Cases, r)
	}

	return result
}

// decodeHex 解码hex字符串
func (f *Fuzzer) decodeHex(hexStr string) (*types.Message, error) {
	if len(hexStr) < 4 {
		return nil, fmt.Errorf("hex too short")
	}

	// 简单解析
	if len(hexStr)%2 != 0 {
		return nil, fmt.Errorf("odd hex length")
	}

	// 尝试解码
	data := make([]byte, len(hexStr)/2)
	for i := 0; i < len(data); i++ {
		var b byte
		_, err := fmt.Sscanf(hexStr[i*2:i*2+2], "%x", &b)
		if err != nil {
			return nil, fmt.Errorf("hex decode error at %d: %w", i, err)
		}
		data[i] = b
	}

	if len(data) < 2 || data[0] != 0x7E || data[len(data)-1] != 0x7E {
		return nil, fmt.Errorf("invalid frame delimiter")
	}

	msg, err := f.codec.Decode(data)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

// mutateHex 随机突变hex字符串
func mutateHex(hexStr string, rng *rand.Rand) string {
	bytes := []byte(hexStr)
	mutations := rng.Intn(3) + 1
	for i := 0; i < mutations; i++ {
		pos := rng.Intn(len(bytes))
		if bytes[pos] >= '0' && bytes[pos] <= '9' || bytes[pos] >= 'A' && bytes[pos] <= 'F' || bytes[pos] >= 'a' && bytes[pos] <= 'f' {
			bytes[pos] = byte('0' + rng.Intn(10))
		}
	}
	return string(bytes)
}

// randomHex 生成随机hex字符串
func randomHex(n int) string {
	bytes := make([]byte, n)
	for i := range bytes {
		bytes[i] = byte('A' + rand.Intn(16))
	}
	return string(bytes)
}
