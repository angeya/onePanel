package main

import (
	"context"
	"fmt"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

/**
 * 打开选择目录对话框
 */
func (a *App) OpenDirectoryDialog(title string) (string, error) {
	dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: title,
	})
	if err != nil {
		return "", fmt.Errorf("打开目录对话框失败: %w", err)
	}
	return dir, nil
}

/**
 * 打开选择文件对话框
 */
func (a *App) OpenFileDialog(title string, filter string) (string, error) {
	filters := []runtime.FileFilter{}
	if filter != "" {
		filters = append(filters, runtime.FileFilter{
			DisplayName: filter,
			Pattern:     filter,
		})
	}

	file, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title:   title,
		Filters: filters,
	})
	if err != nil {
		return "", fmt.Errorf("打开文件对话框失败: %w", err)
	}
	return file, nil
}
