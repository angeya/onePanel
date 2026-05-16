package main

import (
	"runtime"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

var (
	user32hotkey         = syscall.NewLazyDLL("user32.dll")
	procRegisterHotKey   = user32hotkey.NewProc("RegisterHotKey")
	procUnregisterHotKey = user32hotkey.NewProc("UnregisterHotKey")
	procPeekMessageW     = user32hotkey.NewProc("PeekMessageW")
)

const (
	WM_HOTKEY = 0x0312

	MOD_ALT     = 0x0001
	MOD_CONTROL = 0x0002
	MOD_SHIFT   = 0x0004
	MOD_WIN     = 0x0008

	HOTKEY_ID = 1

	VK_O = 0x4F
)

/**
 * HotkeyManager 全局快捷键管理器
 * 通过 Windows RegisterHotKey API 注册全局快捷键
 * 在独立线程中注册和监听快捷键消息，确保同一线程注册和接收
 */
type HotkeyManager struct {
	mu         sync.Mutex
	registered bool
	onActivate func()
	stopCh     chan struct{}
	doneCh     chan struct{}
}

/**
 * 创建 HotkeyManager 实例
 * onActivate: 快捷键触发回调
 */
func NewHotkeyManager(onActivate func()) *HotkeyManager {
	return &HotkeyManager{
		onActivate: onActivate,
		stopCh:     make(chan struct{}),
		doneCh:     make(chan struct{}),
	}
}

/**
 * 启动全局快捷键监听
 * 在独立 goroutine 中注册 Ctrl+Alt+O 并监听 WM_HOTKEY 消息
 * RegisterHotKey 和 PeekMessageW 必须在同一个 OS 线程中调用
 */
func (h *HotkeyManager) Start() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.registered {
		return nil
	}

	h.registered = true
	h.doneCh = make(chan struct{})
	go h.listen()
	return nil
}

/**
 * 在独立线程中注册快捷键并监听消息
 * RegisterHotKey 将 WM_HOTKEY 发送到调用线程的消息队列
 * 因此 PeekMessageW 必须在同一个线程中调用
 */
func (h *HotkeyManager) listen() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret, _, err := procRegisterHotKey.Call(
		0,
		HOTKEY_ID,
		MOD_CONTROL|MOD_ALT,
		VK_O,
	)
	if ret == 0 {
		LogWarn("注册全局快捷键失败: %v", err)
		h.mu.Lock()
		h.registered = false
		h.mu.Unlock()
		close(h.doneCh)
		return
	}

	defer func() {
		procUnregisterHotKey.Call(0, HOTKEY_ID)
		close(h.doneCh)
	}()

	var msg MSG
	for {
		select {
		case <-h.stopCh:
			return
		default:
		}

		ret, _, _ := procPeekMessageW.Call(
			uintptr(unsafe.Pointer(&msg)),
			0, WM_HOTKEY, WM_HOTKEY,
			0x0001, // PM_REMOVE
		)
		if ret != 0 && msg.Message == WM_HOTKEY {
			if h.onActivate != nil {
				h.onActivate()
			}
		} else {
			time.Sleep(50 * time.Millisecond)
		}
	}
}

/**
 * 停止全局快捷键监听
 * 通知监听线程退出，并等待其完成注销
 */
func (h *HotkeyManager) Stop() {
	h.mu.Lock()
	if !h.registered {
		h.mu.Unlock()
		return
	}
	h.registered = false
	h.mu.Unlock()

	close(h.stopCh)
	<-h.doneCh
	h.stopCh = make(chan struct{})
}
