package main

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

/**
 * 时间格式常量
 * 统一数据库中使用的日期时间格式
 */
const (
	DateTimeFormat     = "2006-01-02 15:04:05"
	DateTimeCompactFmt = "20060102150405"
)

/**
 * 获取当前时间的格式化字符串
 * 统一使用 DateTimeFormat 格式
 */
func NowFormatted() string {
	return time.Now().Format(DateTimeFormat)
}

/**
 * 获取当前时间的紧凑格式化字符串
 * 用于生成文件名等场景
 */
func NowCompactFormatted() string {
	return time.Now().Format(DateTimeCompactFmt)
}

/**
 * 获取默认 Shell
 * 如果传入的 shell 为空，则返回 cmd.exe
 */
func DefaultShell(shell string) string {
	if shell == "" {
		return "cmd.exe"
	}
	return shell
}

/**
 * 压缩目录为 zip 文件
 * 将 srcDir 目录及其内容压缩到 zipPath 指定的文件中
 */
func ZipDirectory(srcDir, zipPath string) error {
	file, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer file.Close()

	w := zip.NewWriter(file)
	defer w.Close()

	baseDir := filepath.Base(srcDir)
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		zipEntry := filepath.Join(baseDir, relPath)
		zipEntry = filepath.ToSlash(zipEntry)

		if info.IsDir() {
			_, err := w.Create(zipEntry + "/")
			return err
		}

		fileWriter, err := w.Create(zipEntry)
		if err != nil {
			return err
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(fileWriter, f)
		return err
	})
}

/**
 * 复制目录及其所有内容
 * 递归复制 src 目录到 dst 目录
 */
func CopyDirectory(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, 0755)
		}

		if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
			return err
		}

		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		dstFile, err := os.Create(dstPath)
		if err != nil {
			return err
		}
		defer dstFile.Close()

		_, err = io.Copy(dstFile, srcFile)
		return err
	})
}

/**
 * 清理目录名称中的非法字符
 * 将 Windows 文件系统不允许的字符替换为下划线
 */
func SanitizeDirName(name string) string {
	invalid := []string{"\\", "/", ":", "*", "?", "\"", "<", ">", "|"}
	for _, ch := range invalid {
		name = strings.ReplaceAll(name, ch, "_")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		name = "app_" + NowCompactFormatted()
	}
	return name
}

/**
 * 获取应用数据存储目录
 * 位于可执行文件同级的 data 目录下
 */
func GetDataDir() string {
	exePath, err := os.Executable()
	if err != nil {
		return "./data"
	}
	return filepath.Join(filepath.Dir(exePath), "data")
}
