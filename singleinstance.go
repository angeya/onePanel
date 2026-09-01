package main

import (
	"golang.org/x/sys/windows"
)

/**
 * Windows 单实例支持。
 * 通过命名互斥体保证全局只有一个进程实例；
 * 再次启动时通过命名事件通知已有实例激活窗口后退出。
 */
const (
	singleInstanceMutexName = "Local\\oneWin-single-instance"
	activateEventName       = "Local\\oneWin-activate"
)

// keepMutex 持有单实例互斥体句柄，进程退出时由系统自动释放。
var keepMutex windows.Handle

/**
 * TryAcquireSingleInstance 尝试创建单实例命名互斥体。
 * 返回 true 表示获取成功（首个实例）；返回 false 表示已有实例在运行。
 */
func TryAcquireSingleInstance() bool {
	mutex, err := windows.CreateMutex(nil, false, windows.StringToUTF16Ptr(singleInstanceMutexName))
	if err == windows.ERROR_ALREADY_EXISTS {
		return false
	}
	if err != nil {
		// 创建互斥体异常时不阻断启动，仅记录日志
		LogWarn("创建单实例互斥体失败: %v", err)
		return true
	}
	keepMutex = mutex
	return true
}

/**
 * NotifyExistingInstance 通知已有实例激活其窗口。
 * 若通知事件不存在（如旧版本未内置该功能），返回错误由调用方决定如何处理。
 */
func NotifyExistingInstance() error {
	event, err := windows.OpenEvent(windows.EVENT_MODIFY_STATE, false, windows.StringToUTF16Ptr(activateEventName))
	if err != nil {
		return err
	}
	defer windows.CloseHandle(event)
	return windows.SetEvent(event)
}

/**
 * StartActivateWatcher 启动后台协程监听激活事件。
 * 收到第二次启动的通知时调用 onActivate 回调（通常为显示并聚焦主窗口）。
 */
func StartActivateWatcher(onActivate func()) {
	go func() {
		event, err := windows.CreateEvent(nil, 0, 0, windows.StringToUTF16Ptr(activateEventName))
		if err != nil {
			LogWarn("创建激活事件失败: %v", err)
			return
		}
		for {
			_, err := windows.WaitForSingleObject(event, windows.INFINITE)
			if err != nil {
				LogWarn("等待激活事件失败: %v", err)
				return
			}
			if onActivate != nil {
				onActivate()
			}
		}
	}()
}
