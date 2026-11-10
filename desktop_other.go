//go:build !windows

package main

import (
	"fmt"
	"os"
)

// ensureWebView2Runtime 非 Windows 平台无需检测（WebView2 仅存在于 Windows）
func ensureWebView2Runtime() {}

// showErrorBox 非 Windows 桌面模式一般由命令行启动，输出到 stderr 即可
func showErrorBox(title, text string) {
	fmt.Fprintf(os.Stderr, "%s: %s\n", title, text)
}
