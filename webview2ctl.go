package main

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"syscall"
	"unsafe"

	"github.com/wailsapp/go-webview2/pkg/edge"
)

var (
	cmUser32 = syscall.NewLazyDLL("user32.dll")

	cmProcRegisterWindowMessageW = cmUser32.NewProc("RegisterWindowMessageW")
	cmProcSetWindowLongPtrW      = cmUser32.NewProc("SetWindowLongPtrW")
	cmProcCallWindowProcW        = cmUser32.NewProc("CallWindowProcW")
	cmProcPostMessageW           = cmUser32.NewProc("PostMessageW")
)

/**
 * 上下文菜单控制状态
 * 保存子类化窗口过程所需的运行时数据
 */
var cmState struct {
	sync.Mutex
	chromium    *edge.Chromium
	hwnd        uintptr
	origWndProc uintptr
	msgId       uint32
	enabled     bool
	initialized bool
}

var cmWndProc uintptr

/**
 * 初始化上下文菜单控制
 * 通过子类化 WebView2 主窗口的窗口过程，
 * 使 COM 调用可以在 UI 线程上执行
 * 必须在 WebView2 初始化完成后调用
 */
func InitContextMenuControl(ctx context.Context) error {
	chromium, err := getChromium(ctx)
	if err != nil {
		return err
	}

	cmState.Lock()
	defer cmState.Unlock()

	cmState.chromium = chromium
	cmState.hwnd = *(*uintptr)(unsafe.Pointer(chromium))

	msgName, _ := syscall.UTF16PtrFromString("ONEWIN_SET_CONTEXT_MENU")
	ret, _, _ := cmProcRegisterWindowMessageW.Call(uintptr(unsafe.Pointer(msgName)))
	if ret == 0 {
		return fmt.Errorf("注册窗口消息失败")
	}
	cmState.msgId = uint32(ret)

	cmWndProc = syscall.NewCallback(cmWindowProc)

	origRet, _, _ := cmProcSetWindowLongPtrW.Call(
		cmState.hwnd,
		uintptr(0xFFFFFFFFFFFFFFFC),
		cmWndProc,
	)
	if origRet == 0 {
		return fmt.Errorf("子类化窗口过程失败")
	}
	cmState.origWndProc = origRet
	cmState.initialized = true

	return nil
}

/**
 * 恢复原始窗口过程
 * 在应用退出时调用，确保窗口过程被正确还原
 */
func CleanupContextMenuControl() {
	cmState.Lock()
	defer cmState.Unlock()

	if !cmState.initialized {
		return
	}

	cmProcSetWindowLongPtrW.Call(
		cmState.hwnd,
		uintptr(0xFFFFFFFFFFFFFFFC),
		cmState.origWndProc,
	)
	cmState.initialized = false
}

/**
 * 子类化窗口过程
 * 拦截自定义消息，在 UI 线程上执行 WebView2 COM 调用
 * 其他消息转发给原始窗口过程处理
 */
func cmWindowProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	cmState.Lock()
	msgId := cmState.msgId
	origProc := cmState.origWndProc
	chromium := cmState.chromium
	enabled := cmState.enabled
	cmState.Unlock()

	if msg == msgId && chromium != nil {
		settings, err := chromium.GetSettings()
		if err == nil {
			settings.PutAreDefaultContextMenusEnabled(enabled)
		}
		return 0
	}

	ret, _, _ := cmProcCallWindowProcW.Call(origProc, hwnd, uintptr(msg), wParam, lParam)
	return ret
}

/**
 * 设置 WebView2 默认上下文菜单是否启用
 * 通过 PostMessage 将请求投递到 UI 线程，
 * 由子类化窗口过程在 UI 线程上执行 COM 调用
 */
func SetContextMenuEnabled(ctx context.Context, enabled bool) error {
	cmState.Lock()
	if !cmState.initialized {
		cmState.Unlock()
		return fmt.Errorf("上下文菜单控制未初始化")
	}
	cmState.enabled = enabled
	hwnd := cmState.hwnd
	msgId := cmState.msgId
	cmState.Unlock()

	cmProcPostMessageW.Call(hwnd, uintptr(msgId), 0, 0)
	return nil
}

/**
 * 从 Wails 上下文中获取 *edge.Chromium 对象
 *
 * 开发模式访问链路:
 *   ctx.Value("frontend") -> *DevWebServer -> .Frontend(嵌入接口) -> 解包 -> *windows.Frontend -> .chromium
 *
 * 生产模式访问链路:
 *   ctx.Value("frontend") -> *windows.Frontend -> .chromium
 *
 * 由于 chromium 是未导出字段，使用 unsafe.Pointer 读取
 */
func getChromium(ctx context.Context) (*edge.Chromium, error) {
	if ctx == nil {
		return nil, fmt.Errorf("上下文为空")
	}

	frontendInterface := ctx.Value("frontend")
	if frontendInterface == nil {
		return nil, fmt.Errorf("无法从上下文获取前端接口")
	}

	frontendVal, err := resolveFrontendStruct(frontendInterface)
	if err != nil {
		return nil, err
	}

	chromiumField := frontendVal.FieldByName("chromium")
	if !chromiumField.IsValid() {
		return nil, fmt.Errorf("无法访问 chromium 字段 (类型: %s)", frontendVal.Type().Name())
	}

	chromiumPtr := *(**edge.Chromium)(unsafe.Pointer(chromiumField.UnsafeAddr()))
	if chromiumPtr == nil {
		return nil, fmt.Errorf("chromium 对象为空")
	}

	return chromiumPtr, nil
}

/**
 * 解析前端结构体的 reflect.Value
 * 处理开发模式（DevWebServer）和生产模式（windows.Frontend）两种情况
 */
func resolveFrontendStruct(frontendInterface interface{}) (reflect.Value, error) {
	val := unwrapInterface(reflect.ValueOf(frontendInterface))
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	chromiumField := val.FieldByName("chromium")
	if chromiumField.IsValid() {
		return val, nil
	}

	embeddedFrontend := val.FieldByName("Frontend")
	if embeddedFrontend.IsValid() {
		innerVal := unwrapInterface(embeddedFrontend)
		if innerVal.Kind() == reflect.Ptr {
			innerVal = innerVal.Elem()
		}
		if innerVal.FieldByName("chromium").IsValid() {
			return innerVal, nil
		}
	}

	return reflect.Value{}, fmt.Errorf("无法解析前端结构 (类型: %s)", val.Type().Name())
}

/**
 * 解包接口类型的 reflect.Value
 * 当 reflect.ValueOf(interface{}) 得到的是 Interface 类型的 Value 时，
 * 需要调用 Elem() 获取底层具体的值
 */
func unwrapInterface(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Interface && !v.IsNil() {
		v = v.Elem()
	}
	return v
}
