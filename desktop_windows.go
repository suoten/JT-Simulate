//go:build windows

package main

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"

	"github.com/wailsapp/go-webview2/webviewloader"
)

// ensureWebView2Runtime 桌面模式启动前检测 WebView2 运行时。
// wails 底层（go-webview2）在运行时缺失/异常时会静默 os.Exit(1)，双击运行时
// 没有控制台，表现为"双击无反应"——因此必须在进入 wails 前主动检测并给出可见提示。
func ensureWebView2Runtime() {
	version, err := webviewloader.GetAvailableCoreWebView2BrowserVersionString("")
	if err == nil && version != "" {
		return // 运行时已安装
	}
	detail := "未检测到 WebView2 运行时"
	if err != nil {
		detail = fmt.Sprintf("WebView2 运行时检测失败: %v", err)
	}
	// https://go.microsoft.com/fwlink/?linkid=2124703 为微软官方 Evergreen
	// 独立安装引导程序（WebView2 Runtime Installer）
	showErrorBox("JT-Simulate 需要 WebView2 运行时",
		detail+"。\n\n本程序桌面窗口依赖 Microsoft Edge WebView2 运行时。\n\n"+
			"请下载并安装后重新运行本程序：\nhttps://go.microsoft.com/fwlink/?linkid=2124703\n\n"+
			"（也可改用服务模式：在命令行运行 jt-simulate.exe serve，用浏览器访问提示的地址）")
	os.Exit(1)
}

// showErrorBox 弹出 Windows 消息框（双击运行时没有控制台，必须用 GUI 提示）
func showErrorBox(title, text string) {
	const mbOK = 0x00000000
	const mbIconWarning = 0x00000030
	user32 := syscall.NewLazyDLL("user32.dll")
	messageBoxW := user32.NewProc("MessageBoxW")
	titlePtr, _ := syscall.UTF16PtrFromString(title)
	textPtr, _ := syscall.UTF16PtrFromString(text)
	messageBoxW.Call(0,
		uintptr(unsafe.Pointer(textPtr)),
		uintptr(unsafe.Pointer(titlePtr)),
		uintptr(mbOK|mbIconWarning))
}
