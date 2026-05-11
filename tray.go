package main

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"unsafe"
)

var (
	shell32    = syscall.NewLazyDLL("shell32.dll")
	user32tray = syscall.NewLazyDLL("user32.dll")
	kernel32   = syscall.NewLazyDLL("kernel32.dll")

	procShellNotifyIconW    = shell32.NewProc("Shell_NotifyIconW")
	procCreateWindowExW     = user32tray.NewProc("CreateWindowExW")
	procDefWindowProcW      = user32tray.NewProc("DefWindowProcW")
	procRegisterClassExW    = user32tray.NewProc("RegisterClassExW")
	procDestroyWindow       = user32tray.NewProc("DestroyWindow")
	procPostQuitMessage     = user32tray.NewProc("PostQuitMessage")
	procGetMessageW         = user32tray.NewProc("GetMessageW")
	procTranslateMessage    = user32tray.NewProc("TranslateMessage")
	procDispatchMessageW    = user32tray.NewProc("DispatchMessageW")
	procSetForegroundWindow = user32tray.NewProc("SetForegroundWindow")
	procLoadImageW          = user32tray.NewProc("LoadImageW")
	procPostMessageW        = user32tray.NewProc("PostMessageW")
	procGetModuleHandleW    = kernel32.NewProc("GetModuleHandleW")
	procLoadIconW           = user32tray.NewProc("LoadIconW")
	procCreatePopupMenu     = user32tray.NewProc("CreatePopupMenu")
	procAppendMenuW         = user32tray.NewProc("AppendMenuW")
	procTrackPopupMenu      = user32tray.NewProc("TrackPopupMenu")
	procDestroyMenu         = user32tray.NewProc("DestroyMenu")
	procGetCursorPos        = user32tray.NewProc("GetCursorPos")
	procDestroyIcon         = user32tray.NewProc("DestroyIcon")

	cwUseDefault uintptr
)

func init() {
	n := math.MaxInt32
	cwUseDefault = uintptr(n + 1)
}

const (
	NIM_ADD    = 0x00000000
	NIM_MODIFY = 0x00000001
	NIM_DELETE = 0x00000002

	NIF_MESSAGE = 0x00000001
	NIF_ICON    = 0x00000002
	NIF_TIP     = 0x00000004

	WM_USER      = 0x0400
	WM_TRAYICON  = WM_USER + 1
	WM_LBUTTONUP = 0x0202
	WM_RBUTTONUP = 0x0205
	WM_COMMAND   = 0x0111
	WM_DESTROY   = 0x0002

	IDM_SHOW = 0x1001
	IDM_QUIT = 0x1002

	IMAGE_ICON      = 1
	LR_LOADFROMFILE = 0x00000010
)

type NOTIFYICONDATA struct {
	CbSize           uint32
	HWnd             syscall.Handle
	uID              uint32
	UFlags           uint32
	UCallbackMessage uint32
	HIcon            syscall.Handle
	SzTip            [128]uint16
}

type WNDCLASSEX struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     syscall.Handle
	HIcon         syscall.Handle
	HCursor       syscall.Handle
	HbrBackground syscall.Handle
	LpszMenuName  *uint16
	LpszClassName *uint16
	HIconSm       syscall.Handle
}

type MSG struct {
	HWnd    syscall.Handle
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      struct{ X, Y int32 }
}

/**
 * TrayManager 系统托盘管理器
 * 通过 Windows Shell_NotifyIcon API 实现系统托盘功能
 * 创建隐藏窗口接收托盘消息，支持左键点击显示窗口和右键菜单
 */
type TrayManager struct {
	hWnd        syscall.Handle
	hIcon       syscall.Handle
	nid         NOTIFYICONDATA
	mu          sync.Mutex
	onShow      func()
	onQuit      func()
	initialized bool
}

/**
 * 创建 TrayManager 实例
 * onShow: 左键点击托盘或菜单"显示"回调
 * onQuit: 菜单"退出"回调
 */
func NewTrayManager(onShow func(), onQuit func()) *TrayManager {
	return &TrayManager{
		onShow: onShow,
		onQuit: onQuit,
	}
}

/**
 * 在独立线程中启动托盘
 * 创建隐藏窗口和托盘图标，进入消息循环
 */
func (t *TrayManager) Start() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.initialized {
		return
	}

	go t.messageLoop()
}

/**
 * 托盘消息循环
 * 注册窗口类、创建隐藏窗口、添加托盘图标、进入消息循环
 * 整个循环在一个独立线程中运行
 */
func (t *TrayManager) messageLoop() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	hInstance, _, _ := procGetModuleHandleW.Call(0)

	classNameStr, _ := syscall.UTF16PtrFromString("OneWinTrayClass")
	windowNameStr, _ := syscall.UTF16PtrFromString("OneWinTrayWindow")

	wndClass := WNDCLASSEX{}
	wndClass.CbSize = uint32(unsafe.Sizeof(wndClass))
	wndClass.LpfnWndProc = syscall.NewCallback(t.wndProc)
	wndClass.HInstance = syscall.Handle(hInstance)
	wndClass.LpszClassName = classNameStr

	ret, _, _ := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wndClass)))
	if ret == 0 {
		fmt.Println("注册托盘窗口类失败")
		return
	}

	hWnd, _, _ := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(classNameStr)),
		uintptr(unsafe.Pointer(windowNameStr)),
		0,
		cwUseDefault, cwUseDefault, cwUseDefault, cwUseDefault,
		0, 0, hInstance, 0,
	)
	if hWnd == 0 {
		fmt.Println("创建托盘窗口失败")
		return
	}

	t.hWnd = syscall.Handle(hWnd)

	t.hIcon = t.loadIcon()
	if t.hIcon == 0 {
		fmt.Println("加载托盘图标失败")
	}

	t.nid = NOTIFYICONDATA{}
	t.nid.CbSize = uint32(unsafe.Sizeof(t.nid))
	t.nid.HWnd = t.hWnd
	t.nid.uID = 1
	t.nid.UFlags = NIF_MESSAGE | NIF_ICON | NIF_TIP
	t.nid.UCallbackMessage = WM_TRAYICON
	t.nid.HIcon = t.hIcon
	copy(t.nid.SzTip[:], utf16FromString("oneWin"))

	ret, _, _ = procShellNotifyIconW.Call(NIM_ADD, uintptr(unsafe.Pointer(&t.nid)))
	if ret == 0 {
		fmt.Println("添加托盘图标失败")
	}

	t.mu.Lock()
	t.initialized = true
	t.mu.Unlock()

	var msg MSG
	for {
		ret, _, _ := procGetMessageW.Call(
			uintptr(unsafe.Pointer(&msg)),
			0, 0, 0,
		)
		if ret == 0 || int32(ret) == -1 {
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
	}
}

/**
 * 窗口过程回调
 * 处理托盘图标消息和菜单命令
 */
func (t *TrayManager) wndProc(hWnd syscall.Handle, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case WM_TRAYICON:
		switch lParam {
		case WM_LBUTTONUP:
			if t.onShow != nil {
				t.onShow()
			}
		case WM_RBUTTONUP:
			t.showContextMenu()
		}
	case WM_COMMAND:
		switch wParam {
		case IDM_SHOW:
			if t.onShow != nil {
				t.onShow()
			}
		case IDM_QUIT:
			if t.onQuit != nil {
				t.onQuit()
			}
		}
	case WM_DESTROY:
		procShellNotifyIconW.Call(NIM_DELETE, uintptr(unsafe.Pointer(&t.nid)))
		if t.hIcon != 0 {
			procDestroyIcon.Call(uintptr(t.hIcon))
		}
		procPostQuitMessage.Call(0)
	}
	ret, _, _ := procDefWindowProcW.Call(uintptr(hWnd), uintptr(msg), wParam, lParam)
	return ret
}

/**
 * 显示右键菜单
 * 在托盘图标位置弹出上下文菜单
 */
func (t *TrayManager) showContextMenu() {
	hMenu, _, _ := procCreatePopupMenu.Call()

	showText, _ := syscall.UTF16PtrFromString("显示主窗口")
	quitText, _ := syscall.UTF16PtrFromString("退出")

	procAppendMenuW.Call(hMenu, 0, IDM_SHOW, uintptr(unsafe.Pointer(showText)))
	procAppendMenuW.Call(hMenu, 0x00000800, 0, 0) // MF_SEPARATOR
	procAppendMenuW.Call(hMenu, 0, IDM_QUIT, uintptr(unsafe.Pointer(quitText)))

	var pt struct{ X, Y int32 }
	procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))

	procSetForegroundWindow.Call(uintptr(t.hWnd))

	procTrackPopupMenu.Call(
		hMenu,
		0x0002|0x0010, // TPM_RIGHTALIGN | TPM_BOTTOMALIGN
		uintptr(pt.X), uintptr(pt.Y),
		0, uintptr(t.hWnd), 0,
	)

	procDestroyMenu.Call(hMenu)
}

/**
 * 加载托盘图标
 * 优先加载可执行文件内嵌图标，其次加载 data 目录下的图标文件
 */
func (t *TrayManager) loadIcon() syscall.Handle {
	hInstance, _, _ := procGetModuleHandleW.Call(0)

	hIcon, _, _ := procLoadIconW.Call(hInstance, uintptr(1))
	if hIcon != 0 {
		return syscall.Handle(hIcon)
	}

	iconPath := t.findIconFile()
	if iconPath != "" {
		iconPathPtr, _ := syscall.UTF16PtrFromString(iconPath)
		hIcon, _, _ := procLoadImageW.Call(
			0,
			uintptr(unsafe.Pointer(iconPathPtr)),
			IMAGE_ICON,
			0, 0,
			LR_LOADFROMFILE,
		)
		if hIcon != 0 {
			return syscall.Handle(hIcon)
		}
	}

	hIcon, _, _ = procLoadIconW.Call(0, uintptr(0x7F00)) // IDI_APPLICATION
	return syscall.Handle(hIcon)
}

/**
 * 查找图标文件
 * 依次检查可执行文件目录下的 icon.ico 和 data/icon.ico
 */
func (t *TrayManager) findIconFile() string {
	exePath, err := os.Executable()
	if err != nil {
		return ""
	}
	exeDir := filepath.Dir(exePath)

	candidates := []string{
		filepath.Join(exeDir, "icon.ico"),
		filepath.Join(exeDir, "data", "icon.ico"),
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

/**
 * 移除托盘图标并通知消息循环退出
 * 通过 PostMessageW 向托盘窗口发送 WM_DESTROY 消息
 * 实际的窗口销毁和图标移除由消息循环所在线程处理
 */
func (t *TrayManager) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.initialized {
		return
	}
	if t.hWnd != 0 {
		procPostMessageW.Call(uintptr(t.hWnd), WM_DESTROY, 0, 0)
	}
	t.initialized = false
}

func utf16FromString(s string) []uint16 {
	runes := []rune(s)
	result := make([]uint16, len(runes)+1)
	for i, r := range runes {
		result[i] = uint16(r)
	}
	return result
}
