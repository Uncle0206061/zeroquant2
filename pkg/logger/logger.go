// Package logger 提供统一的日志记录功能
// 支持日志级别控制、日志文件输出、JSON 格式
package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	LevelDebug = "debug"
	LevelInfo = "info"
	LevelWarn = "warn"
	LevelError = "error"
)

var (
	logger       *log.Logger
	level       string
	logFile     *os.File
	mu          sync.Mutex
	isInitialized = false
)

// Init 初始化日志系统
func Init(levelStr, logPath string) {
	if isInitialized {
		return
	}
	level = levelStr

	// 创建日志目录
	if logPath != "" {
		dir := filepath.Dir(logPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create log directory: %v\n", err)
		}
		// 尝试创建日志文件
		logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to open log file: %v\n", err)
			// 降级到 stdout
			logger = log.New(os.Stdout, "", log.LstdFlags)
		} else {
			logger = log.New(io.MultiWriter(os.Stdout, logFile), "", log.LstdFlags)
		}
	} else {
		logger = log.New(os.Stdout, "", log.LstdFlags)
	}

	isInitialized = true
	Info("Logger initialized, level: %s", level)
}

// getLevel 判断日志级别
func checkLevel(levelStr string) bool {
	switch level {
	case LevelDebug:
		return true
	case LevelInfo:
		return levelStr != LevelDebug
	case LevelWarn:
		return levelStr == LevelWarn || levelStr == LevelError
	case LevelError:
		return levelStr == LevelError
	default:
		return true
	}
}

// Debug 调试日志
func Debug(format string, v ...interface{}) {
	if !checkLevel(LevelDebug) {
		return
	}
	logger.Printf("[DEBUG] "+format, v...)
}

// Info 信息日志
func Info(format string, v ...interface{}) {
	if !checkLevel(LevelInfo) {
		return
	}
	logger.Printf("[INFO] "+format, v...)
}

// Warn 警告日志
func Warn(format string, v ...interface{}) {
	if !checkLevel(LevelWarn) {
		return
	}
	logger.Printf("[WARN] "+format, v...)
}

// Error 错误日志
func Error(format string, v ...interface{}) {
	if !checkLevel(LevelError) {
		return
	}
	logger.Printf("[ERROR] "+format, v...)
}

// Fatal 致命错误并退出
func Fatal(format string, v ...interface{}) {
	logger.Printf("[FATAL] "+format, v...)
	os.Exit(1)
}

// GetLogFileName 获取带日期的日志文件名
func GetLogFileName(baseName string) string {
	now := time.Now()
	return fmt.Sprintf("%s.%s.log", baseName, now.Format("2006-01-02"))
}