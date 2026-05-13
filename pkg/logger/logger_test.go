package logger_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/Uncle0206061/zeroquant2/backend/pkg/logger"
)

func TestLoggerInit(t *testing.T) {
	logger.ResetForTest()
	var buf bytes.Buffer
	logger.Init("info", "")
	logger.SetOutput(&buf)
	logger.Info("test")
	if buf.Len() == 0 {
		t.Error("Expected log output after init")
	}
}

func TestLogLevelFiltering(t *testing.T) {
	logger.ResetForTest()
	var buf bytes.Buffer
	logger.Init("error", "")
	logger.SetOutput(&buf)

	// 过滤级别的消息不应该出现
	logger.Debug("debug should not appear")
	logger.Info("info should not appear")
	logger.Warn("warn should not appear")

	output := buf.String()
	if strings.Contains(output, "debug should not appear") {
		t.Error("Debug should be filtered at error level")
	}
	if strings.Contains(output, "info should not appear") {
		t.Error("Info should be filtered at error level")
	}
	if strings.Contains(output, "warn should not appear") {
		t.Error("Warn should be filtered at error level")
	}
	buf.Reset()

	// error 级别消息应该出现
	logger.Error("error should appear")
	output = buf.String()
	if !strings.Contains(output, "error should appear") {
		t.Error("Error should appear at error level")
	}
}

func TestLogOutputFormat(t *testing.T) {
	logger.ResetForTest()
	var buf bytes.Buffer
	logger.Init("info", "")
	logger.SetOutput(&buf)

	logger.Info("test message")

	output := strings.TrimSpace(buf.String())
	if output == "" {
		t.Fatal("Expected log output")
	}

	// 验证是 JSON 格式
	var entry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &entry); err != nil {
		t.Fatalf("Log output is not valid JSON: %s\nError: %v", output, err)
	}

	// 验证必要字段
	if entry["level"] != "info" {
		t.Errorf("Expected level 'info', got '%v'", entry["level"])
	}
	if entry["message"] != "test message" {
		t.Errorf("Expected message 'test message', got '%v'", entry["message"])
	}
	if entry["timestamp"] == "" {
		t.Error("Expected timestamp to be set")
	}
}

func TestLogLevelDebug(t *testing.T) {
	logger.ResetForTest()
	var buf bytes.Buffer
	logger.Init("debug", "")
	logger.SetOutput(&buf)

	logger.Debug("debug message")
	output := buf.String()
	if !strings.Contains(output, "debug message") {
		t.Error("Debug message should appear when level is debug")
	}
}

func TestLogLevelWarn(t *testing.T) {
	logger.ResetForTest()
	var buf bytes.Buffer
	logger.Init("warn", "")
	logger.SetOutput(&buf)

	logger.Info("info should be filtered")
	logger.Warn("warn should appear")
	logger.Error("error should appear")

	output := buf.String()
	if strings.Contains(output, "info should be filtered") {
		t.Error("Info should be filtered at warn level")
	}
	if !strings.Contains(output, "warn should appear") {
		t.Error("Warn should appear at warn level")
	}
	if !strings.Contains(output, "error should appear") {
		t.Error("Error should appear at warn level")
	}
}

func TestLogLevels(t *testing.T) {
	logger.ResetForTest()
	var buf bytes.Buffer
	logger.Init("debug", "")
	logger.SetOutput(&buf)

	logger.Debug("debug-msg")
	logger.Info("info-msg")
	logger.Warn("warn-msg")
	logger.Error("error-msg")

	output := buf.String()
	for _, level := range []string{"debug", "info", "warn", "error"} {
		if !strings.Contains(output, level+"-msg") {
			t.Errorf("Expected %s message in output", level)
		}
	}
}
