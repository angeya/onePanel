package main

import (
	"fmt"
	"runtime"
	"strings"
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
)

/**
 * HotkeyConfig 全局快捷键配置。
 */
type HotkeyConfig struct {
	Modifiers []string `json:"modifiers"`
	Key       string   `json:"key"`
}

/**
 * HotkeyManager 全局快捷键管理器
 * 通过 Windows RegisterHotKey API 注册全局快捷键
 * 在独立线程中注册和监听快捷键消息，确保同一线程注册和接收
 */
type HotkeyManager struct {
	mu         sync.Mutex
	registered bool
	config     HotkeyConfig
	onActivate func()
	stopCh     chan struct{}
	doneCh     chan struct{}
}

/**
 * 创建 HotkeyManager 实例
 * onActivate: 快捷键触发回调
 */
func NewHotkeyManager(config HotkeyConfig, onActivate func()) *HotkeyManager {
	if config.Key == "" {
		config = DefaultHotkeyConfig()
	}
	return &HotkeyManager{
		config:     config,
		onActivate: onActivate,
		stopCh:     make(chan struct{}),
		doneCh:     make(chan struct{}),
	}
}

/**
 * DefaultHotkeyConfig 返回默认快捷键配置 Ctrl+Alt+O。
 */
func DefaultHotkeyConfig() HotkeyConfig {
	return HotkeyConfig{
		Modifiers: []string{"ctrl", "alt"},
		Key:       "O",
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

	mod, vk, err := parseHotkeyConfig(h.config)
	if err != nil {
		LogWarn("解析全局快捷键配置失败: %v", err)
		h.mu.Lock()
		h.registered = false
		h.mu.Unlock()
		close(h.doneCh)
		return
	}

	ret, _, err := procRegisterHotKey.Call(
		0,
		HOTKEY_ID,
		uintptr(mod),
		uintptr(vk),
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

/**
 * parseHotkeyConfig 将 HotkeyConfig 解析为 RegisterHotKey 需要的修饰符和虚拟键码。
 */
func parseHotkeyConfig(config HotkeyConfig) (uint32, uint32, error) {
	var mod uint32
	for _, m := range config.Modifiers {
		switch m {
		case "alt":
			mod |= MOD_ALT
		case "ctrl":
			mod |= MOD_CONTROL
		case "shift":
			mod |= MOD_SHIFT
		case "win":
			mod |= MOD_WIN
		default:
			return 0, 0, fmt.Errorf("不支持的修饰键: %s", m)
		}
	}
	if mod == 0 {
		return 0, 0, fmt.Errorf("至少需要一个修饰键")
	}

	vk, err := keyToVK(config.Key)
	if err != nil {
		return 0, 0, err
	}
	return mod, vk, nil
}

/**
 * keyToVK 将单个字符或功能键名称转换为虚拟键码。
 * 支持 A-Z、0-9 以及 F1-F24。
 */
func keyToVK(key string) (uint32, error) {
	if key == "" {
		return 0, fmt.Errorf("快捷键按键不能为空")
	}

	key = strings.ToUpper(key)

	if len(key) == 1 {
		c := key[0]
		if c >= 'A' && c <= 'Z' {
			return uint32(c), nil
		}
		if c >= '0' && c <= '9' {
			return uint32(c), nil
		}
		return 0, fmt.Errorf("不支持的按键: %s", key)
	}

	if strings.HasPrefix(key, "F") {
		var n int
		if _, err := fmt.Sscanf(key, "F%d", &n); err == nil && n >= 1 && n <= 24 {
			return 0x6F + uint32(n), nil
		}
	}

	return 0, fmt.Errorf("不支持的按键: %s", key)
}
