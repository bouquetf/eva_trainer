package services

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"eva_trainer/pkg/config"

	"github.com/gin-gonic/gin"
)

func TestLogLevels(t *testing.T) {
	var buf bytes.Buffer
	logger := &loggerImpl{level: LogLevelDebug, format: LogFormatText, out: &buf}

	logger.Debug("debug msg")
	logger.Info("info msg")
	logger.Warn("warn msg")
	logger.Error("error msg")

	output := buf.String()
	if !strings.Contains(output, "[DEBUG] debug msg") {
		t.Error("missing debug")
	}
	if !strings.Contains(output, "[INFO] info msg") {
		t.Error("missing info")
	}
	if !strings.Contains(output, "[WARN] warn msg") {
		t.Error("missing warn")
	}
	if !strings.Contains(output, "[ERROR] error msg") {
		t.Error("missing error")
	}
}

func TestLogLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	logger := &loggerImpl{level: LogLevelWarn, format: LogFormatText, out: &buf}

	logger.Debug("debug msg")
	logger.Info("info msg")
	logger.Warn("warn msg")
	logger.Error("error msg")

	output := buf.String()
	if strings.Contains(output, "debug msg") {
		t.Error("debug should be filtered")
	}
	if strings.Contains(output, "info msg") {
		t.Error("info should be filtered")
	}
	if !strings.Contains(output, "warn msg") {
		t.Error("warn should appear")
	}
	if !strings.Contains(output, "error msg") {
		t.Error("error should appear")
	}
}

func TestLogWithContext(t *testing.T) {
	var buf bytes.Buffer
	logger := &loggerImpl{level: LogLevelInfo, format: LogFormatText, out: &buf}

	ctxLogger := logger.With("component", "test", "user_id", 42)
	ctxLogger.Info("contextual log")

	output := buf.String()
	if !strings.Contains(output, "component=test") {
		t.Error("missing component")
	}
	if !strings.Contains(output, "user_id=42") {
		t.Error("missing user_id")
	}
	if !strings.Contains(output, "contextual log") {
		t.Error("missing message")
	}
}

func TestLogFormatText(t *testing.T) {
	var buf bytes.Buffer
	logger := &loggerImpl{level: LogLevelInfo, format: LogFormatText, out: &buf}

	logger.Info("test message", "key1", "val1", "key2", 2)

	output := buf.String()
	if !strings.Contains(output, "[INFO]") {
		t.Error("missing level")
	}
	if !strings.Contains(output, "test message") {
		t.Error("missing message")
	}
	if !strings.Contains(output, "key1=val1") {
		t.Error("missing key1=val1")
	}
	if !strings.Contains(output, "key2=2") {
		t.Error("missing key2=2")
	}

	if strings.HasPrefix(output, "{") {
		t.Error("text format should not be JSON")
	}
}

func TestLogFormatJSON(t *testing.T) {
	var buf bytes.Buffer
	logger := &loggerImpl{level: LogLevelInfo, format: LogFormatJSON, out: &buf}

	logger.Info("json test", "key1", "val1", "num", 42)

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if result["level"] != "INFO" {
		t.Errorf("level = %v", result["level"])
	}
	if result["message"] != "json test" {
		t.Errorf("message = %v", result["message"])
	}
	if result["key1"] != "val1" {
		t.Errorf("key1 = %v", result["key1"])
	}
	if result["num"] != float64(42) {
		t.Errorf("num = %v", result["num"])
	}
	if _, ok := result["time"]; !ok {
		t.Error("missing time field")
	}
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input   string
		want    LogLevel
		wantErr bool
	}{
		{"debug", LogLevelDebug, false},
		{"info", LogLevelInfo, false},
		{"warn", LogLevelWarn, false},
		{"warning", LogLevelWarn, false},
		{"error", LogLevelError, false},
		{"invalid", LogLevelInfo, true},
		{"DEBUG", LogLevelDebug, false},
		{"INFO", LogLevelInfo, false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseLogLevel(tt.input)
			if tt.wantErr && err == nil {
				t.Error("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseLogFormat(t *testing.T) {
	tests := []struct {
		input   string
		want    LogFormat
		wantErr bool
	}{
		{"auto", LogFormatAuto, false},
		{"text", LogFormatText, false},
		{"json", LogFormatJSON, false},
		{"invalid", LogFormatAuto, true},
		{"JSON", LogFormatJSON, false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseLogFormat(tt.input)
			if tt.wantErr && err == nil {
				t.Error("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoggerConcurrency(t *testing.T) {
	var buf bytes.Buffer
	logger := &loggerImpl{level: LogLevelDebug, format: LogFormatText, out: &buf}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			logger.Info("concurrent", "goroutine", i)
		}(i)
	}
	wg.Wait()

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 10 {
		t.Errorf("expected 10 log lines, got %d", len(lines))
	}
}

func TestRequestLogger(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var buf bytes.Buffer
	logger := &loggerImpl{level: LogLevelInfo, format: LogFormatText, out: &buf}

	r := gin.New()
	r.Use(RequestLogger(logger))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	output := buf.String()
	if !strings.Contains(output, "method=GET") {
		t.Error("missing method")
	}
	if !strings.Contains(output, "path=/test") {
		t.Error("missing path")
	}
	if !strings.Contains(output, "status=200") {
		t.Error("missing status")
	}
	if !strings.Contains(output, "duration=") {
		t.Error("missing duration")
	}
}

func TestNewLoggerFromConfig(t *testing.T) {
	cfg := &config.AppConfig{
		Log: config.LogConfig{
			Level:  "debug",
			Format: "json",
		},
	}
	logger := NewLoggerFromConfig(cfg)
	if logger == nil {
		t.Fatal("expected non-nil logger")
	}
	if logger.Level() != LogLevelDebug {
		t.Errorf("level = %v", logger.Level())
	}

	cfg.Log.Level = "invalid"
	logger2 := NewLoggerFromConfig(cfg)
	if logger2.Level() != LogLevelInfo {
		t.Errorf("expected INFO fallback, got %v", logger2.Level())
	}
}

func TestNewLogger(t *testing.T) {
	logger := NewLogger(LogLevelInfo, LogFormatText)
	if logger == nil {
		t.Fatal("expected non-nil")
	}
	if logger.Level() != LogLevelInfo {
		t.Errorf("level = %v", logger.Level())
	}
}

func TestSetLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := &loggerImpl{level: LogLevelError, format: LogFormatText, out: &buf}

	logger.SetLevel(LogLevelDebug)
	if logger.Level() != LogLevelDebug {
		t.Error("level not updated")
	}

	logger.Debug("should appear now")
	if !strings.Contains(buf.String(), "should appear now") {
		t.Error("debug message should appear after SetLevel")
	}
}
