package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

/**
 * 日志保留天数
 * 超过此天数的日志文件将在启动时被清理
 */
const LogRetentionDays = 7

/**
 * 日志文件日期格式
 * 用于日志文件名中的日期部分
 */
const LogDateFormat = "2006-01-02"

var (
	infoLogger  *log.Logger
	warnLogger  *log.Logger
	errorLogger *log.Logger
	logFile     *os.File
)

/**
 * 获取日志目录路径
 * 位于可执行文件同级的 logs 目录下
 */
func GetLogsDir() string {
	exePath, err := os.Executable()
	if err != nil {
		return "./logs"
	}
	return filepath.Join(filepath.Dir(exePath), "logs")
}

/**
 * 初始化日志系统
 * 创建日志目录、打开日志文件、初始化分级日志记录器
 * 同时清理过期的日志文件
 */
func InitLogger() error {
	logsDir := GetLogsDir()
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return fmt.Errorf("创建日志目录失败: %w", err)
	}

	CleanOldLogs(logsDir)

	logFileName := fmt.Sprintf("onewin_%s.log", time.Now().Format(LogDateFormat))
	logPath := filepath.Join(logsDir, logFileName)

	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("打开日志文件失败: %w", err)
	}

	logFile = f

	multiWriter := io.MultiWriter(os.Stdout, f)

	infoLogger = log.New(multiWriter, "[INFO]  ", log.Ldate|log.Ltime|log.Lshortfile)
	warnLogger = log.New(multiWriter, "[WARN]  ", log.Ldate|log.Ltime|log.Lshortfile)
	errorLogger = log.New(multiWriter, "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile)

	return nil
}

/**
 * 关闭日志文件
 * 在应用退出时调用，确保缓冲区数据刷盘
 */
func CloseLogger() {
	if logFile != nil {
		logFile.Close()
		logFile = nil
	}
}

/**
 * 清理过期的日志文件
 * 删除修改时间超过 LogRetentionDays 天的 .log 文件
 */
func CleanOldLogs(logsDir string) {
	entries, err := os.ReadDir(logsDir)
	if err != nil {
		return
	}

	cutoff := time.Now().AddDate(0, 0, -LogRetentionDays)

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".log") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoff) {
			filePath := filepath.Join(logsDir, entry.Name())
			os.Remove(filePath)
		}
	}
}

/**
 * 记录 INFO 级别日志
 */
func LogInfo(format string, args ...interface{}) {
	if infoLogger != nil {
		infoLogger.Printf(format, args...)
	}
}

/**
 * 记录 WARN 级别日志
 */
func LogWarn(format string, args ...interface{}) {
	if warnLogger != nil {
		warnLogger.Printf(format, args...)
	}
}

/**
 * 记录 ERROR 级别日志
 */
func LogError(format string, args ...interface{}) {
	if errorLogger != nil {
		errorLogger.Printf(format, args...)
	}
}
