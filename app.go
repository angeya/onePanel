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
 * 构建 PowerShell 命令并执行，统一处理 UTF-8 编码输出
 * 解决中文 Windows 下 PowerShell 默认使用 GBK 编码导致中文路径乱码的问题
 */
func (a *App) runPowerShell(psScript string) (string, error) {
	fullScript := "[Console]::OutputEncoding = [System.Text.Encoding]::UTF8; " + psScript

	cmd := exec.Command(ResolveShellPath("powershell"), "-NoProfile", "-NonInteractive", "-Command", fullScript)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("%w, stderr: %s", err, stderr.String())
	}

	result := strings.TrimSpace(stdout.String())
	return result, nil
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

	result, err := a.runPowerShell(psScript)
	if err != nil {
		return "", fmt.Errorf("打开目录对话框失败: %w", err)
	}
	if result == "" {
		return "", nil
	}
	return result, nil
}

/**
 * 打开保存文件对话框（使用 PowerShell 实现）
 * 让用户选择导出文件的保存路径
 * filter 参数应为 Windows Forms 标准格式: "显示名称|文件模式"，如 "ZIP 文件 (*.zip)|*.zip"
 */
func (a *App) SaveFileDialog(title string, defaultName string, filter string) (string, error) {
	if title == "" {
		title = "保存文件"
	}

	filterStr := "所有文件 (*.*)|*.*"
	if filter != "" {
		filterStr = filter
	}

	psScript := fmt.Sprintf(`
Add-Type -AssemblyName System.Windows.Forms
$saveFileDialog = New-Object System.Windows.Forms.SaveFileDialog
$saveFileDialog.Title = '%s'
$saveFileDialog.FileName = '%s'
$saveFileDialog.Filter = '%s'
$saveFileDialog.OverwritePrompt = $true
if ($saveFileDialog.ShowDialog() -eq [System.Windows.Forms.DialogResult]::OK) {
    Write-Output $saveFileDialog.FileName
} else {
    Write-Output ""
}`, title, defaultName, filterStr)

	result, err := a.runPowerShell(psScript)
	if err != nil {
		return "", fmt.Errorf("打开保存文件对话框失败: %w", err)
	}
	if result == "" {
		return "", nil
	}
	return result, nil
}

/**
 * 打开选择文件对话框（使用 PowerShell 实现，避免 Wails COM 对话框崩溃问题）
 * filter 参数应为 Windows Forms 标准格式: "显示名称|文件模式"，如 "ZIP 文件 (*.zip)|*.zip"
 */
func (a *App) OpenFileDialog(title string, filter string) (string, error) {
	if title == "" {
		title = "选择文件"
	}

	filterStr := "所有文件 (*.*)|*.*"
	if filter != "" {
		filterStr = filter
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

	result, err := a.runPowerShell(psScript)
	if err != nil {
		return "", fmt.Errorf("打开文件对话框失败: %w", err)
	}
	if result == "" {
		return "", nil
	}
	return result, nil
}
