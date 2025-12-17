package logger

import (
	"fmt"
	"log"
	"os"
	"strings"
)

// Logger provides leveled, structured logging with contextual information.
type Logger struct {
	l      *log.Logger
	debug  bool
	fields map[string]any
}

// New constructs a logger writing to stdout with timestamps.
func New() *Logger {
	return &Logger{
		l:      log.New(os.Stdout, "", log.LstdFlags),
		debug:  false,
		fields: make(map[string]any),
	}
}

// SetDebug enables or disables debug-level logging.
func (l *Logger) SetDebug(enabled bool) {
	l.debug = enabled
}

// WithFields returns a new logger instance with additional contextual fields.
func (l *Logger) WithFields(fields map[string]any) *Logger {
	newFields := make(map[string]any)
	for k, v := range l.fields {
		newFields[k] = v
	}
	for k, v := range fields {
		newFields[k] = v
	}
	return &Logger{
		l:      l.l,
		debug:  l.debug,
		fields: newFields,
	}
}

// Debug logs debug-level messages (only if debug mode is enabled).
func (l *Logger) Debug(msg string, kv ...any) {
	if !l.debug {
		return
	}
	l.log("DEBUG", msg, kv...)
}

// Info logs informational messages.
func (l *Logger) Info(msg string, kv ...any) {
	l.log("INFO", msg, kv...)
}

// Warn logs warning messages.
func (l *Logger) Warn(msg string, kv ...any) {
	l.log("WARN", msg, kv...)
}

// Error logs error messages.
func (l *Logger) Error(msg string, kv ...any) {
	l.log("ERROR", msg, kv...)
}

// log formats and outputs a log message with level, message, fields, and key-value pairs.
func (l *Logger) log(level, msg string, kv ...any) {
	var parts []string
	parts = append(parts, fmt.Sprintf("[%s]", level), msg)

	// Add structured fields
	if len(l.fields) > 0 {
		var fieldParts []string
		for k, v := range l.fields {
			fieldParts = append(fieldParts, fmt.Sprintf("%s=%v", k, v))
		}
		parts = append(parts, strings.Join(fieldParts, " "))
	}

	// Add key-value pairs from arguments
	if len(kv) > 0 {
		for i := 0; i < len(kv); i += 2 {
			if i+1 < len(kv) {
				parts = append(parts, fmt.Sprintf("%v=%v", kv[i], kv[i+1]))
			} else {
				parts = append(parts, fmt.Sprintf("%v", kv[i]))
			}
		}
	}

	l.l.Println(strings.Join(parts, " "))
}



