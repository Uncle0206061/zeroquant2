// Package logger 提供统一的日志记录功能
// 支持：结构化 JSON 格式、request_id、日志级别控制、文件轮转
package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// 日志级别常量
const (
	LevelDebug = "debug"
	LevelInfo  = "info"
	LevelWarn  = "warn"
	LevelError = "error"
)

// levelOrder 级别排序（数值越大越严格）
var levelOrder = map[string]int{
	LevelDebug: 0,
	LevelInfo:  1,
	LevelWarn:  2,
	LevelError: 3,
}

// logEntry 结构化日志条目
type logEntry struct {
	Level     string `json:"level"`               // 日志级别
	Timestamp string `json:"timestamp"`           // ISO8601 时间戳
	Message   string `json:"message"`             // 日志消息
	RequestID string `json:"request_id,omitempty"` // 请求ID（可选）
	Caller    string `json:"caller,omitempty"`     // 调用方（可选）
}

var (
	writer        io.Writer
	currentLevel  string
	logFile       *os.File
	mu            sync.Mutex
	isInitialized bool
)

// Init 初始化日志系统
// levelStr: debug/info/warn/error
// logPath: 日志文件路径（空则仅输出到 stdout）
func Init(levelStr, logPath string) {
	if isInitialized {
		return
	}
	currentLevel = levelStr

	// 创建日志目录
	if logPath != "" {
		dir := filepath.Dir(logPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create log directory: %v\n", err)
		}
		f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to open log file: %v\n", err)
			writer = os.Stdout
		} else {
			logFile = f
			writer = io.MultiWriter(os.Stdout, f)
		}
	} else {
		writer = os.Stdout
	}

	isInitialized = true
	Info("Logger initialized, level: %s", levelStr)
}

// SetOutput 设置输出目标（用于测试）
func SetOutput(w io.Writer) {
	mu.Lock()
	defer mu.Unlock()
	writer = w
}

// checkLevel 检查日志级别是否允许输出
func checkLevel(level string) bool {
	current, ok1 := levelOrder[currentLevel]
	target, ok2 := levelOrder[level]
	if !ok1 || !ok2 {
		return true // 未知级别默认输出
	}
	return target >= current
}

// writeLog 写入结构化日志
func writeLog(level, requestID, format string, v ...interface{}) {
	if !isInitialized {
		return
	}
	if !checkLevel(level) {
		return
	}

	entry := logEntry{
		Level:     level,
		Timestamp: time.Now().Format(time.RFC3339Nano),
		Message:   fmt.Sprintf(format, v...),
		RequestID: requestID,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		// 降级：纯文本输出
		fmt.Fprintf(writer, "[%s] %s %s\n", level, entry.Timestamp, entry.Message)
		return
	}

	mu.Lock()
	defer mu.Unlock()
	fmt.Fprintln(writer, string(data))
}

// ============ 通用日志函数（无 requestID）============

// Debug 调试日志
func Debug(format string, v ...interface{}) {
	writeLog(LevelDebug, "", format, v...)
}

// Info 信息日志
func Info(format string, v ...interface{}) {
	writeLog(LevelInfo, "", format, v...)
}

// Warn 警告日志
func Warn(format string, v ...interface{}) {
	writeLog(LevelWarn, "", format, v...)
}

// Error 错误日志
func Error(format string, v ...interface{}) {
	writeLog(LevelError, "", format, v...)
}

// Fatal 致命错误并退出
func Fatal(format string, v ...interface{}) {
	writeLog(LevelError, "", format, v...)
	os.Exit(1)
}

// ============ 带 requestID 的日志函数 ============

// DebugR 带 requestID 的调试日志
func DebugR(requestID, format string, v ...interface{}) {
	writeLog(LevelDebug, requestID, format, v...)
}

// InfoR 带 requestID 的信息日志
func InfoR(requestID, format string, v ...interface{}) {
	writeLog(LevelInfo, requestID, format, v...)
}

// WarnR 带 requestID 的警告日志
func WarnR(requestID, format string, v ...interface{}) {
	writeLog(LevelWarn, requestID, format, v...)
}

// ErrorR 带 requestID 的错误日志
func ErrorR(requestID, format string, v ...interface{}) {
	writeLog(LevelError, requestID, format, v...)
}

// GetLogFileName 获取带日期的日志文件名
func GetLogFileName(baseName string) string {
	return fmt.Sprintf("%s.%s.log", baseName, time.Now().Format("2006-01-02"))
}
