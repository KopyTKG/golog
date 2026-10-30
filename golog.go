package golog

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

type flags int

const (
	DISABLED flags = iota
	ENABLED
)

type level int

const (
	SILENT level = iota
	ERROR
	WARN
	INFO
	DEBUG
)

var LOGLEVEL = SILENT
var CLILOG = ENABLED
var isLogFileEnabled = DISABLED
var outputFormat = "text" // "text" or "json"

var cliWriter io.Writer = os.Stdout
var fileWriter io.Writer = nil

// Fields represents structured logging fields
type Fields map[string]interface{}

// Logger represents a named logger with optional fields
type Logger struct {
	name   string
	fields Fields
	mu     sync.RWMutex
}

// Global default logger
var defaultLogger = &Logger{
	name:   "",
	fields: make(Fields),
}

// contextKey is the key type for storing logger in context
type contextKey struct{}

// SetOutputFormat sets the output format ("text" or "json")
func SetOutputFormat(format string) {
	if format == "json" || format == "text" {
		outputFormat = format
	}
}

// Named creates a new named logger
func Named(name string) *Logger {
	return &Logger{
		name:   name,
		fields: make(Fields),
	}
}

// WithFields creates a logger with additional fields
func WithFields(fields Fields) *Logger {
	l := &Logger{
		name:   defaultLogger.name,
		fields: make(Fields),
	}
	for k, v := range fields {
		l.fields[k] = v
	}
	return l
}

// WithFields adds fields to a named logger
func (l *Logger) WithFields(fields Fields) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()

	newLogger := &Logger{
		name:   l.name,
		fields: make(Fields),
	}

	// Copy existing fields
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}

	// Add new fields
	for k, v := range fields {
		newLogger.fields[k] = v
	}

	return newLogger
}

// WithContext adds logger to context
func (l *Logger) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextKey{}, l)
}

// FromContext retrieves logger from context
func FromContext(ctx context.Context) *Logger {
	if l, ok := ctx.Value(contextKey{}).(*Logger); ok {
		return l
	}
	return defaultLogger
}

// formatMessage formats a log message with fields
func formatMessage(level string, msg string, name string, fields Fields) string {
	now := time.Now()

	if outputFormat == "json" {
		// JSON format
		jsonData := map[string]interface{}{
			"time":  now.Format(time.RFC3339),
			"level": strings.ToLower(level),
			"msg":   msg,
		}

		if name != "" {
			jsonData["component"] = name
		}

		for k, v := range fields {
			jsonData[k] = v
		}

		data, _ := json.Marshal(jsonData)
		return string(data) + "\n"
	}

	// Text format (with colors for CLI)
	timeStr := now.Format("15:04:05")

	// Build fields string
	fieldsStr := ""
	if len(fields) > 0 {
		parts := []string{}
		for k, v := range fields {
			parts = append(parts, fmt.Sprintf("%s=%v", k, v))
		}
		fieldsStr = " " + strings.Join(parts, " ")
	}

	// Add component name
	componentStr := ""
	if name != "" {
		componentStr = fmt.Sprintf(" [%s]", name)
	}

	return fmt.Sprintf("%s %s%s %s%s", timeStr, level, componentStr, msg, fieldsStr)
}

// getLevelColor returns ANSI color code for level
func getLevelColor(level string) string {
	switch level {
	case "INFO":
		return "\u001b[36m" // Cyan
	case "OK":
		return "\u001b[32m" // Green
	case "SYSW", "APPW", "WARN":
		return "\u001b[33m" // Yellow
	case "ERRO":
		return "\u001b[31m" // Red
	case "DEBG":
		return "\u001b[35m" // Magenta
	case "FATAL":
		return "\u001b[31;1m" // Bold red
	default:
		return "\u001b[37m" // White
	}
}

// writeLog writes to CLI and/or file
func writeLog(level string, msg string, name string, fields Fields) {
	if outputFormat == "json" {
		// JSON format (same for CLI and file)
		output := formatMessage(level, msg, name, fields)
		if CLILOG == ENABLED {
			fmt.Fprint(cliWriter, output)
		}
		if isLogFileEnabled == ENABLED && fileWriter != nil {
			fmt.Fprint(fileWriter, output)
		}
	} else {
		// Text format
		now := time.Now().Format("15:04:05")

		// Build fields string
		fieldsStr := ""
		if len(fields) > 0 {
			parts := []string{}
			for k, v := range fields {
				parts = append(parts, fmt.Sprintf("%s=%v", k, v))
			}
			fieldsStr = " " + strings.Join(parts, " ")
		}

		// Component name
		componentStr := ""
		if name != "" {
			componentStr = fmt.Sprintf(" [%s]", name)
		}

		// CLI output (with colors)
		if CLILOG == ENABLED {
			timeColor := "\u001b[38;5;244m"
			levelColor := getLevelColor(level)
			reset := "\u001b[0m"

			cliText := fmt.Sprintf("%s%s%s %s%s%s%s %s%s\n",
				timeColor, now, reset,
				levelColor, level, reset,
				componentStr,
				msg, fieldsStr)
			fmt.Fprint(cliWriter, cliText)
		}

		// File output (no colors)
		if isLogFileEnabled == ENABLED && fileWriter != nil {
			fileText := fmt.Sprintf("%s %s%s %s%s\n",
				now, level, componentStr, msg, fieldsStr)
			fmt.Fprint(fileWriter, fileText)
		}
	}
}

func EnableLogFile(path string) error {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	fileWriter = file
	isLogFileEnabled = ENABLED
	return nil
}

func DisableLogFile() {
	if fileWriter != nil {
		if f, ok := fileWriter.(*os.File); ok {
			f.Close()
		}
	}
	fileWriter = nil
	isLogFileEnabled = DISABLED
}

/* BACKWARD COMPATIBLE - NEW LINE FUNCTIONS */

func Info(msg any) {
	if LOGLEVEL >= INFO {
		writeLog("INFO", fmt.Sprint(msg), "", nil)
	}
}

func Success(msg any) {
	if LOGLEVEL >= INFO {
		writeLog("OK  ", fmt.Sprint(msg), "", nil)
	}
}

func Warn(msg any) {
	if LOGLEVEL >= WARN {
		writeLog("WARN", fmt.Sprint(msg), "", nil)
	}
}

func SYSWarn(msg any) {
	if LOGLEVEL >= WARN {
		writeLog("SYSW", fmt.Sprint(msg), "", nil)
	}
}

func APPWarn(msg any) {
	if LOGLEVEL >= WARN {
		writeLog("APPW", fmt.Sprint(msg), "", nil)
	}
}

func Error(msg any) {
	if LOGLEVEL >= ERROR {
		writeLog("ERRO", fmt.Sprint(msg), "", nil)
	}
}

func Debug(msg any) {
	if LOGLEVEL >= DEBUG {
		writeLog("DEBG", fmt.Sprint(msg), "", nil)
	}
}

func Fatal(msg any) {
	writeLog("FATAL", fmt.Sprint(msg), "", nil)
	os.Exit(1)
}

/* BACKWARD COMPATIBLE - FORMATTED FUNCTIONS */

func Infof(format string, args ...any) {
	if LOGLEVEL >= INFO {
		msg := fmt.Sprintf(format, args...)
		writeLog("INFO", msg, "", nil)
	}
}

func Successf(format string, args ...any) {
	if LOGLEVEL >= INFO {
		msg := fmt.Sprintf(format, args...)
		writeLog("OK  ", msg, "", nil)
	}
}

func Warnf(format string, args ...any) {
	if LOGLEVEL >= WARN {
		msg := fmt.Sprintf(format, args...)
		writeLog("WARN", msg, "", nil)
	}
}

func SYSWarnf(format string, args ...any) {
	if LOGLEVEL >= WARN {
		msg := fmt.Sprintf(format, args...)
		writeLog("SYSW", msg, "", nil)
	}
}

func APPWarnf(format string, args ...any) {
	if LOGLEVEL >= WARN {
		msg := fmt.Sprintf(format, args...)
		writeLog("APPW", msg, "", nil)
	}
}

func Errorf(format string, args ...any) {
	if LOGLEVEL >= ERROR {
		msg := fmt.Sprintf(format, args...)
		writeLog("ERRO", msg, "", nil)
	}
}

func Debugf(format string, args ...any) {
	if LOGLEVEL >= DEBUG {
		msg := fmt.Sprintf(format, args...)
		writeLog("DEBG", msg, "", nil)
	}
}

func Fatalf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	writeLog("FATAL", msg, "", nil)
	os.Exit(1)
}

/* NEW - CONTEXT-AWARE FUNCTIONS */

func InfoContext(ctx context.Context, msg string) {
	if LOGLEVEL >= INFO {
		l := FromContext(ctx)
		l.Info(msg)
	}
}

func WarnContext(ctx context.Context, msg string) {
	if LOGLEVEL >= WARN {
		l := FromContext(ctx)
		l.Warn(msg)
	}
}

func ErrorContext(ctx context.Context, msg string) {
	if LOGLEVEL >= ERROR {
		l := FromContext(ctx)
		l.Error(msg)
	}
}

func DebugContext(ctx context.Context, msg string) {
	if LOGLEVEL >= DEBUG {
		l := FromContext(ctx)
		l.Debug(msg)
	}
}

/* NEW - LOGGER METHODS (for named loggers and structured logging) */

func (l *Logger) Info(msg string) {
	if LOGLEVEL >= INFO {
		l.mu.RLock()
		defer l.mu.RUnlock()
		writeLog("INFO", msg, l.name, l.fields)
	}
}

func (l *Logger) Infof(format string, args ...any) {
	if LOGLEVEL >= INFO {
		l.mu.RLock()
		defer l.mu.RUnlock()
		msg := fmt.Sprintf(format, args...)
		writeLog("INFO", msg, l.name, l.fields)
	}
}

func (l *Logger) Success(msg string) {
	if LOGLEVEL >= INFO {
		l.mu.RLock()
		defer l.mu.RUnlock()
		writeLog("OK  ", msg, l.name, l.fields)
	}
}

func (l *Logger) Successf(format string, args ...any) {
	if LOGLEVEL >= INFO {
		l.mu.RLock()
		defer l.mu.RUnlock()
		msg := fmt.Sprintf(format, args...)
		writeLog("OK  ", msg, l.name, l.fields)
	}
}

func (l *Logger) Warn(msg string) {
	if LOGLEVEL >= WARN {
		l.mu.RLock()
		defer l.mu.RUnlock()
		writeLog("WARN", msg, l.name, l.fields)
	}
}

func (l *Logger) Warnf(format string, args ...any) {
	if LOGLEVEL >= WARN {
		l.mu.RLock()
		defer l.mu.RUnlock()
		msg := fmt.Sprintf(format, args...)
		writeLog("WARN", msg, l.name, l.fields)
	}
}

func (l *Logger) Error(msg string) {
	if LOGLEVEL >= ERROR {
		l.mu.RLock()
		defer l.mu.RUnlock()
		writeLog("ERRO", msg, l.name, l.fields)
	}
}

func (l *Logger) Errorf(format string, args ...any) {
	if LOGLEVEL >= ERROR {
		l.mu.RLock()
		defer l.mu.RUnlock()
		msg := fmt.Sprintf(format, args...)
		writeLog("ERRO", msg, l.name, l.fields)
	}
}

func (l *Logger) ErrorWithStack(err error) {
	if LOGLEVEL >= ERROR && err != nil {
		l.mu.RLock()
		defer l.mu.RUnlock()

		// Get stack trace
		buf := make([]byte, 4096)
		n := runtime.Stack(buf, false)
		stack := string(buf[:n])

		msg := fmt.Sprintf("%v\n%s", err, stack)
		writeLog("ERRO", msg, l.name, l.fields)
	}
}

func (l *Logger) Debug(msg string) {
	if LOGLEVEL >= DEBUG {
		l.mu.RLock()
		defer l.mu.RUnlock()
		writeLog("DEBG", msg, l.name, l.fields)
	}
}

func (l *Logger) Debugf(format string, args ...any) {
	if LOGLEVEL >= DEBUG {
		l.mu.RLock()
		defer l.mu.RUnlock()
		msg := fmt.Sprintf(format, args...)
		writeLog("DEBG", msg, l.name, l.fields)
	}
}

func (l *Logger) Fatal(msg string) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	writeLog("FATAL", msg, l.name, l.fields)
	os.Exit(1)
}

func (l *Logger) Fatalf(format string, args ...any) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	msg := fmt.Sprintf(format, args...)
	writeLog("FATAL", msg, l.name, l.fields)
	os.Exit(1)
}

/* UTILITY FUNCTIONS */

// ErrorWithStack logs error with stack trace (global function)
func ErrorWithStack(err error) {
	if LOGLEVEL >= ERROR && err != nil {
		// Get stack trace
		buf := make([]byte, 4096)
		n := runtime.Stack(buf, false)
		stack := string(buf[:n])

		msg := fmt.Sprintf("%v\n%s", err, stack)
		writeLog("ERRO", msg, "", nil)
	}
}
