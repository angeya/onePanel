package main

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
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
 * 打开选择目录对话框（使用 PowerShell 实现，避免 Wails COM 对话框崩溃问题）
 */
func (a *App) OpenDirectoryDialog(title string) (string, error) {
	if title == "" {
		title = "选择目录"
	}

	psScript := fmt.Sprintf(`
Add-Type -AssemblyName System.Windows.Forms
$folderBrowser = New-Object System.Windows.Forms.FolderBrowserDialog
$folderBrowser.Description = '%s'
$folderBrowser.ShowNewFolderButton = $true
if ($folderBrowser.ShowDialog() -eq [System.Windows.Forms.DialogResult]::OK) {
    Write-Output $folderBrowser.SelectedPath
} else {
    Write-Output ""
}`, title)

	cmd := exec.Command(ResolveShellPath("powershell"), "-NoProfile", "-NonInteractive", "-Command", psScript)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("打开目录对话框失败: %w, stderr: %s", err, stderr.String())
	}

	result := strings.TrimSpace(stdout.String())
	if result == "" {
		return "", nil
	}
	return result, nil
}

/**
 * 打开选择文件对话框（使用 PowerShell 实现，避免 Wails COM 对话框崩溃问题）
 */
func (a *App) OpenFileDialog(title string, filter string) (string, error) {
	if title == "" {
		title = "选择文件"
	}

	filterStr := "所有文件 (*.*)|*.*"
	if filter != "" {
		filterStr = fmt.Sprintf("%s|%s", filter, filter)
	}

	psScript := fmt.Sprintf(`
Add-Type -AssemblyName System.Windows.Forms
$openFileDialog = New-Object System.Windows.Forms.OpenFileDialog
$openFileDialog.Title = '%s'
$openFileDialog.Filter = '%s'
$openFileDialog.Multiselect = $false
if ($openFileDialog.ShowDialog() -eq [System.Windows.Forms.DialogResult]::OK) {
    Write-Output $openFileDialog.FileName
} else {
    Write-Output ""
}`, title, filterStr)

	cmd := exec.Command(ResolveShellPath("powershell"), "-NoProfile", "-NonInteractive", "-Command", psScript)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("打开文件对话框失败: %w, stderr: %s", err, stderr.String())
	}

	result := strings.TrimSpace(stdout.String())
	if result == "" {
		return "", nil
	}
	return result, nil
}
