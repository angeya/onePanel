package main

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

/**
 * App 应用核心结构
 * 持有 Wails 上下文和数据库引用
 * 提供窗口控制、文件对话框等前端可调用方法
 */
type App struct {
	ctx context.Context
	db  *Database
}

/**
 * 创建 App 实例
 * 注入 Database 依赖
 */
func NewApp(db *Database) *App {
	return &App{db: db}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

/**
 * 显示主窗口
 * 将窗口从隐藏状态恢复为可见，并置于前台
 */
func (a *App) ShowWindow() {
	runtime.WindowShow(a.ctx)
}

/**
 * 隐藏主窗口
 * 将窗口隐藏到系统托盘，不退出应用
 */
func (a *App) HideWindow() {
	runtime.WindowHide(a.ctx)
}

/**
 * 退出应用
 * 通知前端即将关闭，前端完成清理后真正退出
 */
func (a *App) QuitApp() {
	runtime.Quit(a.ctx)
}

/**
 * 获取关闭行为设置
 * 返回 "tray"（最小化到托盘）或 "close"（直接退出）
 * 如果未设置则返回空字符串，前端据此判断是否为首次关闭
 */
func (a *App) GetCloseAction() string {
	val, _ := a.db.GetConfig("close_action")
	return val
}

/**
 * 设置关闭行为
 * action: "tray"（最小化到托盘）或 "close"（直接退出）
 */
func (a *App) SetCloseAction(action string) error {
	return a.db.SetConfig("close_action", action)
}

/**
 * 打开选择目录对话框
 * 使用 Wails Runtime 原生 API，直接调用操作系统对话框
 */
func (a *App) OpenDirectoryDialog(title string) (string, error) {
	if title == "" {
		title = "选择目录"
	}

	result, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: title,
	})
	if err != nil {
		return "", err
	}
	return result, nil
}

/**
 * 打开保存文件对话框
 * 让用户选择导出文件的保存路径
 * filter 参数格式: "显示名称|文件模式"，如 "ZIP 文件 (*.zip)|*.zip"
 * 多个过滤器用 "|" 分隔: "ZIP 文件|*.zip|所有文件|*.*"
 */
func (a *App) SaveFileDialog(title string, defaultName string, filter string) (string, error) {
	if title == "" {
		title = "保存文件"
	}

	filters := parseFilters(filter)

	result, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           title,
		DefaultFilename: defaultName,
		Filters:         filters,
	})
	if err != nil {
		return "", err
	}
	return result, nil
}

/**
 * 打开选择文件对话框
 * filter 参数格式: "显示名称|文件模式"，如 "HTML 文件 (*.html;*.htm)|*.html;*.htm"
 * 多个过滤器用 "|" 分隔: "HTML 文件|*.html;*.htm|所有文件|*.*"
 */
func (a *App) OpenFileDialog(title string, filter string) (string, error) {
	if title == "" {
		title = "选择文件"
	}

	filters := parseFilters(filter)

	result, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title:   title,
		Filters: filters,
	})
	if err != nil {
		return "", err
	}
	return result, nil
}

/**
 * 将 Windows Forms 格式的过滤器字符串解析为 Wails Runtime 的 FileFilter 切片
 * 输入格式: "显示名称1|模式1|显示名称2|模式2"
 * 如: "ZIP 文件 (*.zip)|*.zip|所有文件 (*.*)|*.*"
 */
func parseFilters(filter string) []runtime.FileFilter {
	if filter == "" {
		return nil
	}

	parts := splitFilterParts(filter)
	if len(parts) < 2 {
		return nil
	}

	var filters []runtime.FileFilter
	for i := 0; i+1 < len(parts); i += 2 {
		filters = append(filters, runtime.FileFilter{
			DisplayName: parts[i],
			Pattern:     parts[i+1],
		})
	}
	return filters
}

/**
 * 按管道符拆分过滤器字符串
 * 处理过滤器中可能出现的管道符转义
 */
func splitFilterParts(filter string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(filter); i++ {
		if filter[i] == '|' {
			parts = append(parts, filter[start:i])
			start = i + 1
		}
	}
	if start < len(filter) {
		parts = append(parts, filter[start:])
	}
	return parts
}
