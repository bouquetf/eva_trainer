package services

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"eva_trainer/pkg/config"

	"github.com/gin-gonic/gin"
)

type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

var logLevelNames = map[LogLevel]string{
	LogLevelDebug: "DEBUG",
	LogLevelInfo:  "INFO",
	LogLevelWarn:  "WARN",
	LogLevelError: "ERROR",
}

type LogFormat int

const (
	LogFormatAuto LogFormat = iota
	LogFormatText
	LogFormatJSON
)

type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	With(args ...any) Logger
	SetLevel(level LogLevel)
	Level() LogLevel
}

type loggerImpl struct {
	mu     sync.Mutex
	level  LogLevel
	format LogFormat
	ctx    []any
	out    io.Writer
}

func NewLogger(level LogLevel, format LogFormat) Logger {
	return &loggerImpl{
		level:  level,
		format: format,
		ctx:    nil,
		out:    os.Stdout,
	}
}

func NewLoggerFromConfig(cfg *config.AppConfig) Logger {
	level, err := ParseLogLevel(cfg.Log.Level)
	if err != nil {
		level = LogLevelInfo
	}

	format, err := ParseLogFormat(cfg.Log.Format)
	if err != nil {
		format = LogFormatAuto
	}

	if format == LogFormatAuto {
		if cfg.Server.GinMode == "release" {
			format = LogFormatJSON
		} else {
			format = LogFormatText
		}
	}

	return &loggerImpl{
		level:  level,
		format: format,
		out:    os.Stdout,
	}
}

func ParseLogLevel(s string) (LogLevel, error) {
	switch strings.ToLower(s) {
	case "debug":
		return LogLevelDebug, nil
	case "info":
		return LogLevelInfo, nil
	case "warn", "warning":
		return LogLevelWarn, nil
	case "error":
		return LogLevelError, nil
	default:
		return LogLevelInfo, fmt.Errorf("unknown log level: %s", s)
	}
}

func ParseLogFormat(s string) (LogFormat, error) {
	switch strings.ToLower(s) {
	case "auto":
		return LogFormatAuto, nil
	case "text":
		return LogFormatText, nil
	case "json":
		return LogFormatJSON, nil
	default:
		return LogFormatAuto, fmt.Errorf("unknown log format: %s", s)
	}
}

func (l *loggerImpl) Debug(msg string, args ...any) {
	l.log(LogLevelDebug, msg, args...)
}

func (l *loggerImpl) Info(msg string, args ...any) {
	l.log(LogLevelInfo, msg, args...)
}

func (l *loggerImpl) Warn(msg string, args ...any) {
	l.log(LogLevelWarn, msg, args...)
}

func (l *loggerImpl) Error(msg string, args ...any) {
	l.log(LogLevelError, msg, args...)
}

func (l *loggerImpl) With(args ...any) Logger {
	newCtx := make([]any, 0, len(l.ctx)+len(args))
	newCtx = append(newCtx, l.ctx...)
	newCtx = append(newCtx, args...)

	return &loggerImpl{
		level:  l.level,
		format: l.format,
		ctx:    newCtx,
		out:    l.out,
	}
}

func (l *loggerImpl) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

func (l *loggerImpl) Level() LogLevel {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.level
}

func (l *loggerImpl) isEnabled(level LogLevel) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return level >= l.level
}

func (l *loggerImpl) log(level LogLevel, msg string, args ...any) {
	if !l.isEnabled(level) {
		return
	}

	merged := make([]any, 0, len(l.ctx)+len(args))
	merged = append(merged, l.ctx...)
	merged = append(merged, args...)

	var line string
	if l.format == LogFormatJSON {
		line = l.formatJSON(level, msg, merged)
	} else {
		line = l.formatText(level, msg, merged)
	}

	l.write(line)
}

func (l *loggerImpl) formatText(level LogLevel, msg string, args []any) string {
	var b strings.Builder
	b.WriteString(time.Now().Format("2006/01/02 15:04:05"))
	b.WriteString(" [")
	b.WriteString(logLevelNames[level])
	b.WriteString("] ")
	b.WriteString(msg)

	for i := 0; i < len(args); i += 2 {
		b.WriteString(" ")
		if key, ok := args[i].(string); ok {
			b.WriteString(key)
			b.WriteString("=")
			if i+1 < len(args) {
				fmt.Fprint(&b, args[i+1])
			}
		} else {
			fmt.Fprint(&b, args[i])
		}
	}

	return b.String()
}

func (l *loggerImpl) formatJSON(level LogLevel, msg string, args []any) string {
	fields := map[string]any{
		"time":    time.Now().Format(time.RFC3339Nano),
		"level":   logLevelNames[level],
		"message": msg,
	}

	for i := 0; i < len(args); i += 2 {
		if key, ok := args[i].(string); ok && i+1 < len(args) {
			fields[key] = args[i+1]
		}
	}

	data, err := json.Marshal(fields)
	if err != nil {
		return fmt.Sprintf(`{"time":"%s","level":"ERROR","message":"failed to marshal log: %s"}`, time.Now().Format(time.RFC3339Nano), err.Error())
	}

	return string(data)
}

func (l *loggerImpl) write(line string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	_, _ = fmt.Fprintln(l.out, line)
}

func RequestLogger(logger Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		clientIP := c.ClientIP()

		c.Next()

		status := c.Writer.Status()
		duration := time.Since(start)

		logger.With(
			"method", c.Request.Method,
			"path", path,
			"status", status,
			"duration", duration.String(),
			"client_ip", clientIP,
		).Info("request")
	}
}
