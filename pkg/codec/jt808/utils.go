package jt808

import "fmt"

// Escape 转义处理
func Escape(data []byte) []byte {
	result := make([]byte, 0, len(data)*2)
	for _, b := range data {
		switch b {
		case 0x7E:
			result = append(result, 0x7D, 0x02)
		case 0x7D:
			result = append(result, 0x7D, 0x01)
		default:
			result = append(result, b)
		}
	}
	return result
}

// Unescape 反转义处理
func Unescape(data []byte) ([]byte, error) {
	result := make([]byte, 0, len(data))
	i := 0
	for i < len(data) {
		if data[i] == 0x7D {
			if i+1 >= len(data) {
				return nil, fmt.Errorf("Unescape: trailing 0x7D at position %d without escape byte", i)
			}
			switch data[i+1] {
			case 0x02:
				result = append(result, 0x7E)
			case 0x01:
				result = append(result, 0x7D)
			default:
				result = append(result, data[i], data[i+1])
			}
			i += 2
		} else {
			result = append(result, data[i])
			i++
		}
	}
	return result, nil
}

// SplitByDelimiter 按分隔符切分消息
func SplitByDelimiter(data []byte) [][]byte {
	var messages [][]byte
	start := -1

	for i := 0; i < len(data); i++ {
		if data[i] == 0x7E {
			if start == -1 {
				start = i
			} else {
				frameLen := i - start + 1
				if frameLen > 2 && frameLen <= MaxFrameSize {
					msg := make([]byte, frameLen)
					copy(msg, data[start:i+1])
					messages = append(messages, msg)
				}
				start = i
			}
		}
	}

	return messages
}

// WrapWithDelimiter 添加首尾分隔符
func WrapWithDelimiter(data []byte) []byte {
	result := make([]byte, 0, len(data)+2)
	result = append(result, 0x7E)
	result = append(result, data...)
	result = append(result, 0x7E)
	return result
}

// StripDelimiter 去除首尾分隔符
func StripDelimiter(data []byte) []byte {
	if len(data) < 2 {
		return data
	}
	if data[0] == 0x7E && data[len(data)-1] == 0x7E {
		result := make([]byte, len(data)-2)
		copy(result, data[1:len(data)-1])
		return result
	}
	return data
}

// CalcChecksum 计算校验码
func CalcChecksum(data []byte) byte {
	var xor byte
	for _, b := range data {
		xor ^= b
	}
	return xor
}

// BCDToString BCD转字符串（不剥前导零）
func BCDToString(bcd []byte) (string, error) {
	result := make([]byte, 0, len(bcd)*2)
	var firstErr error
	for i, b := range bcd {
		high := b >> 4
		low := b & 0x0F
		if high > 9 || low > 9 {
			if firstErr == nil {
				firstErr = fmt.Errorf("BCDToString: invalid BCD byte 0x%02X at position %d", b, i)
			}
		}
		result = append(result, high+'0', low+'0')
	}
	return string(result), firstErr
}

// BCDToStringSafe BCD转字符串（忽略错误）
func BCDToStringSafe(bcd []byte) string {
	s, _ := BCDToString(bcd)
	return s
}

// StringToBCD 字符串转BCD
// s: 数字字符串
// targetLen: 目标BCD字节数（如6=12位数字）
// 当输入长度不足时左补零；当输入长度超出时取最后 targetLen*2 位
func StringToBCD(s string, targetLen int) ([]byte, error) {
	if targetLen <= 0 {
		targetLen = 6
	}
	// 过滤非数字字符
	filtered := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] >= '0' && s[i] <= '9' {
			filtered = append(filtered, s[i])
		}
	}
	if len(filtered) == 0 {
		return nil, fmt.Errorf("StringToBCD: no digit characters in input %q", s)
	}
	s = string(filtered)
	targetChars := targetLen * 2
	// 不足时左补零
	for len(s) < targetChars {
		s = "0" + s
	}
	// 超出时取最后 N 位（保留低位数字）
	if len(s) > targetChars {
		s = s[len(s)-targetChars:]
	}
	bcd := make([]byte, targetLen)
	for i := 0; i < targetLen; i++ {
		high := s[i*2] - '0'
		low := s[i*2+1] - '0'
		bcd[i] = (high << 4) | low
	}
	return bcd, nil
}

// StringToBCD6 字符串转6字节BCD（用于终端手机号，12位数字）
func StringToBCD6(s string) ([]byte, error) {
	return StringToBCD(s, 6)
}

// trimNull 去除尾部空字节
func trimNull(data []byte) string {
	end := len(data)
	for end > 0 && data[end-1] == 0 {
		end--
	}
	return string(data[:end])
}
