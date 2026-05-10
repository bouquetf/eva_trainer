package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"eva_trainer/internal/services"
	"eva_trainer/pkg/apierror"

	"github.com/gin-gonic/gin"
)

type testLoggerImpl struct {
	buf *bytes.Buffer
	ctx []any
}

func (l *testLoggerImpl) log(msg string, args ...any) {
	fmt.Fprint(l.buf, msg)
	for _, v := range l.ctx {
		fmt.Fprintf(l.buf, " %v", v)
	}
	for _, v := range args {
		fmt.Fprintf(l.buf, " %v", v)
	}
	fmt.Fprintln(l.buf)
}

func (l *testLoggerImpl) Debug(msg string, args ...any) { l.log(msg, args...) }
func (l *testLoggerImpl) Info(msg string, args ...any)  { l.log(msg, args...) }
func (l *testLoggerImpl) Warn(msg string, args ...any)  { l.log(msg, args...) }
func (l *testLoggerImpl) Error(msg string, args ...any) { l.log(msg, args...) }
func (l *testLoggerImpl) With(args ...any) services.Logger {
	newCtx := make([]any, len(l.ctx)+len(args))
	copy(newCtx, l.ctx)
	copy(newCtx[len(l.ctx):], args)
	return &testLoggerImpl{buf: l.buf, ctx: newCtx}
}
func (l *testLoggerImpl) SetLevel(level services.LogLevel) {}
func (l *testLoggerImpl) Level() services.LogLevel         { return services.LogLevelDebug }

func TestRecoveryMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var buf bytes.Buffer
	logger := &testLoggerImpl{buf: &buf}

	r := gin.New()
	r.Use(RecoveryMiddleware(logger))
	r.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/panic", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}

	var resp apierror.APIError
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Code != apierror.ErrCodeInternal {
		t.Errorf("Code = %q, want %q", resp.Code, apierror.ErrCodeInternal)
	}
	if resp.Message != "internal server error" {
		t.Errorf("Message = %q, want %q", resp.Message, "internal server error")
	}

	logOutput := buf.String()
	if !strings.Contains(logOutput, "recovered from panic") {
		t.Errorf("log output should contain 'recovered from panic', got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "test panic") {
		t.Errorf("log output should contain 'test panic', got: %s", logOutput)
	}
}

func TestRecoveryMiddleware_CustomPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var buf bytes.Buffer
	logger := &testLoggerImpl{buf: &buf}

	r := gin.New()
	r.Use(RecoveryMiddleware(logger))
	r.GET("/panic", func(c *gin.Context) {
		panic("custom string panic")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/panic", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}

	var resp apierror.APIError
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Code != apierror.ErrCodeInternal {
		t.Errorf("Code = %q, want %q", resp.Code, apierror.ErrCodeInternal)
	}
	if resp.Message != "internal server error" {
		t.Errorf("Message = %q, want %q", resp.Message, "internal server error")
	}

	logOutput := buf.String()
	if !strings.Contains(logOutput, "recovered from panic") {
		t.Errorf("log output should contain 'recovered from panic'")
	}
	if !strings.Contains(logOutput, "custom string panic") {
		t.Errorf("log output should contain 'custom string panic'")
	}
}

func TestRecoveryMiddleware_NoPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var buf bytes.Buffer
	logger := &testLoggerImpl{buf: &buf}

	r := gin.New()
	r.Use(RecoveryMiddleware(logger))
	r.GET("/ok", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ok", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp["status"] != "ok" {
		t.Errorf("expected status=ok, got %v", resp["status"])
	}
}
