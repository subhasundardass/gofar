// Package logger provides a leveled, structured logger for GoFar.
//
// The Logger wraps the standard library log package with log levels and
// structured key/value fields. In development mode output is human-readable
// (with optional ANSI colors); in production mode output is JSON so it can
// be ingested by log aggregators.
//
// Example:
//
//	log := logger.New(logger.LevelInfo, false)
//	log.Info("server started")
//	log.With("user_id", 42).Warn("suspicious request")
//	log.Errorf("db query failed: %v", err)
//
// A zero-value Logger is NOT safe to use. Always create via New() or Default().
package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"time"
)

// =============================================================================
// Level
// =============================================================================

// Level represents a logging severity.
type Level int

const (
	LevelDebug Level = iota // verbose debugging details
	LevelInfo               // general operational information
	LevelWarn               // non-fatal anomalies
	LevelError              // errors requiring attention
	LevelFatal              // unrecoverable errors – os.Exit(1) is called after log
)

// String returns the upper-case name of the level. It is used by both the
// human and JSON encoders so output is consistent across modes.
func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// =============================================================================
// ANSI palette
// =============================================================================
//
// A small, opinionated palette. We keep all codes as constants so the hot
// path (the formatHuman function below) does not allocate strings for the
// common cases.

const (
	ansiReset     = "\x1b[0m"
	ansiBold      = "\x1b[1m"
	ansiDim       = "\x1b[2m"
	ansiUnderline = "\x1b[4m"

	// Foregrounds
	ansiBlack   = "\x1b[30m"
	ansiRed     = "\x1b[31m"
	ansiGreen   = "\x1b[32m"
	ansiYellow  = "\x1b[33m"
	ansiBlue    = "\x1b[34m"
	ansiMagenta = "\x1b[35m"
	ansiCyan    = "\x1b[36m"
	ansiWhite   = "\x1b[37m"

	// Bright foregrounds
	ansiBrightBlack   = "\x1b[90m"
	ansiBrightRed     = "\x1b[91m"
	ansiBrightGreen   = "\x1b[92m"
	ansiBrightYellow  = "\x1b[93m"
	ansiBrightBlue    = "\x1b[94m"
	ansiBrightMagenta = "\x1b[95m"
	ansiBrightCyan    = "\x1b[96m"
	ansiBrightWhite   = "\x1b[97m"

	// Backgrounds
	ansiBgRed     = "\x1b[41m"
	ansiBgYellow  = "\x1b[43m"
	ansiBgBlue    = "\x1b[44m"
	ansiBgMagenta = "\x1b[45m"
	ansiBgCyan    = "\x1b[46m"
	ansiBgGray    = "\x1b[100m"
)

// =============================================================================
// Logger
// =============================================================================

// Logger writes structured log entries to an io.Writer.
type Logger struct {
	out      io.Writer
	level    Level
	json     bool
	color    bool
	colorFor map[Level]string // foreground for the level tag
	fields   map[string]any
}

// New creates a Logger writing to stderr at the given minimum level.
// If jsonOutput is true each entry is marshalled as a JSON object (suitable
// for production log aggregators); otherwise entries are pretty-printed.
func New(level Level, jsonOutput bool) *Logger {
	return &Logger{
		out:      os.Stderr,
		level:    level,
		json:     jsonOutput,
		color:    shouldColor(jsonOutput, os.Getenv("NO_COLOR") == "", isTTY(os.Stderr)),
		colorFor: defaultLevelPalette(),
		fields:   make(map[string]any),
	}
}

// Default returns a development-friendly logger writing to stderr at DEBUG
// level with human-readable output and ANSI colors enabled.
func Default() *Logger {
	l := New(LevelDebug, false)
	l.color = true
	return l
}

// shouldColor decides whether to enable ANSI colors. It follows the
// standard "NO_COLOR" convention (https://no-color.org) and also disables
// colors automatically when the output is not a TTY (e.g. redirected to
// a file or piped to another process).
func shouldColor(jsonMode, noColorEnv, isTerminal bool) bool {
	if jsonMode {
		return false
	}
	if !noColorEnv {
		return false
	}
	if !isTerminal {
		return false
	}
	return true
}

// SetOutput changes the destination writer. Useful in tests.
//
// NOTE: SetOutput does NOT re-evaluate the color decision. Color is a
// user-facing policy (decided once in New / SetColor) and is independent
// from where bytes happen to be written. Tests that redirect to a
// bytes.Buffer and still want color can call SetColor(true) explicitly.
func (l *Logger) SetOutput(w io.Writer) { l.out = w }

// SetLevel changes the minimum level at runtime.
func (l *Logger) SetLevel(level Level) { l.level = level }

// SetColor enables/disables ANSI colors for the level tag in human mode.
func (l *Logger) SetColor(enabled bool) { l.color = enabled }

// With returns a new Logger that includes the given key/value fields in every
// subsequent entry. The original Logger is unchanged (immutable fields).
//
//	reqLog := log.With("request_id", reqID, "user_id", userID)
//	reqLog.Info("request started")
func (l *Logger) With(kvs ...any) *Logger {
	merged := make(map[string]any, len(l.fields)+len(kvs)/2)
	for k, v := range l.fields {
		merged[k] = v
	}
	for i := 0; i+1 < len(kvs); i += 2 {
		if key, ok := kvs[i].(string); ok {
			merged[key] = kvs[i+1]
		}
	}
	return &Logger{
		out:      l.out,
		level:    l.level,
		json:     l.json,
		color:    l.color,
		colorFor: l.colorFor,
		fields:   merged,
	}
}

// =============================================================================
// TTY detection (kept tiny — we only need it for the auto-color decision)
// =============================================================================
//
// We avoid pulling in golang.org/x/term just for one call. Instead, we do
// a build-tag-gated syscall in isatty_unix.go / isatty_windows.go and
// fall back to "false" (no color) on every other platform.

// fdOf extracts the underlying file descriptor from w when possible. It
// returns ok=false for writers that don't expose a file descriptor
// (e.g. bytes.Buffer in tests) — in that case we conservatively return
// false from isTTY, so test output stays free of ANSI noise.
func fdOf(w io.Writer) (uintptr, bool) {
	type fd interface {
		Fd() uintptr
	}
	if f, ok := w.(fd); ok {
		return f.Fd(), true
	}
	return 0, false
}

// isTTY reports whether w is a terminal. The platform-specific
// implementation lives in isatty_unix.go (Linux, macOS, *BSD) and
// isatty_windows.go; every other platform gets the no-color fallback
// defined in isatty_other.go.
func isTTY(w io.Writer) bool {
	fd, ok := fdOf(w)
	if !ok {
		return false
	}
	return isatty(fd)
}

// =============================================================================
// Encoding helpers
// =============================================================================

// defaultLevelPalette returns the foreground color used for each level tag
// when the level itself is rendered in human mode.
func defaultLevelPalette() map[Level]string {
	return map[Level]string{
		LevelDebug: ansiBrightBlack,
		LevelInfo:  ansiBrightCyan,
		LevelWarn:  ansiBrightYellow,
		LevelError: ansiBrightRed,
		LevelFatal: ansiBrightWhite,
	}
}

// paint returns s wrapped in the given ANSI code if the logger has color
// enabled, or s unchanged otherwise. It centralises the enabled-check so
// every helper below reads cleanly.
func (l *Logger) paint(code, s string) string {
	if !l.color {
		return s
	}
	return code + s + ansiReset
}

// levelPadded returns the level name padded to 5 characters so log
// columns line up vertically in the human-readable output:
//
//	[INFO ]  ...
//	[WARN ]  ...
//	[ERROR]  ...
func levelPadded(l Level) string {
	name := l.String()
	if len(name) >= 5 {
		return name
	}
	return name + strings.Repeat(" ", 5-len(name))
}

// formatValue renders a value in a way that is safe and pretty in plain
// text: strings are quoted (so empty/space-only values are visible), times
// use RFC3339, errors print as quoted messages, complex values
// (map/slice/struct) are JSON-encoded inline, and the rest falls back to
// %v. Colors are applied to the value based on its dynamic type.
func (l *Logger) formatValue(v any) string {
	switch x := v.(type) {
	case string:
		return l.paint(ansiYellow, fmt.Sprintf("%q", x))
	case []byte:
		return l.paint(ansiYellow, fmt.Sprintf("%q", string(x)))
	case bool:
		if x {
			return l.paint(ansiMagenta, "true")
		}
		return l.paint(ansiMagenta, "false")
	case error:
		return l.paint(ansiRed, fmt.Sprintf("%q", x.Error()))
	case time.Time:
		return l.paint(ansiCyan, x.Format(time.RFC3339))
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64:
		return l.paint(ansiGreen, fmt.Sprintf("%v", v))
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Map, reflect.Slice, reflect.Array, reflect.Struct:
		b, err := json.Marshal(v)
		if err == nil {
			return l.paint(ansiCyan, string(b))
		}
	}
	return fmt.Sprintf("%v", v)
}

// formatHuman renders a complete log line in the "pretty" mode. Output
// shape (with colors stripped for clarity):
//
//	2025/06/09 22:37:10 [INFO ] handler.go:51
//	    ↳ server started
//	    • request_id="abc-123"  user_id=42
func (l *Logger) formatHuman(level Level, msg, file string, line int, fields map[string]any) string {
	var b strings.Builder

	// 1. Header line: timestamp (dim gray), level tag (level color, bold),
	//    caller (dim gray).
	ts := l.paint(ansiBrightBlack, time.Now().Format("2006/01/02 15:04:05"))
	tag := l.paint(ansiBold+l.colorFor[level], "["+levelPadded(level)+"]")
	caller := l.paint(ansiDim, fmt.Sprintf("%s:%d", file, line))

	b.WriteString(ts)
	b.WriteByte(' ')
	b.WriteString(tag)
	b.WriteByte(' ')
	b.WriteString(caller)
	b.WriteByte('\n')

	// 2. Message line, indented with an arrow whose color matches the level.
	arrow := l.paint(l.colorFor[level], "↳")
	b.WriteString("    ")
	b.WriteString(arrow)
	b.WriteByte(' ')
	b.WriteString(l.paint(ansiBold+ansiBrightWhite, msg))
	b.WriteByte('\n')

	// 3. Fields, one bullet row. Sorted by key for stable, grep-friendly
	//    output (Go map iteration is otherwise randomised).
	if len(fields) > 0 {
		keys := make([]string, 0, len(fields))
		for k := range fields {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		bullet := l.paint(l.colorFor[level], "•")
		b.WriteString("    ")
		b.WriteString(bullet)
		b.WriteByte(' ')
		for i, k := range keys {
			if i > 0 {
				b.WriteString("  ")
			}
			b.WriteString(l.paint(ansiBrightCyan, k))
			b.WriteByte('=')
			b.WriteString(l.formatValue(fields[k]))
		}
		b.WriteByte('\n')
	}

	return b.String()
}

func (l *Logger) log(level Level, msg string) {
	if level < l.level {
		return
	}
	_, file, line, _ := runtime.Caller(2)
	// Trim full path to just package/file.go
	if idx := strings.LastIndex(file, "/"); idx >= 0 {
		file = file[idx+1:]
	}

	if l.json {
		entry := map[string]any{
			"time":   time.Now().UTC().Format(time.RFC3339),
			"level":  levelNames(level),
			"msg":    msg,
			"caller": fmt.Sprintf("%s:%d", file, line),
		}
		for k, v := range l.fields {
			entry[k] = v
		}
		b, _ := json.Marshal(entry)
		fmt.Fprintf(l.out, "%s\n", b)
	} else {
		fmt.Fprint(l.out, l.formatHuman(level, msg, file, line, l.fields))
	}

	if level == LevelFatal {
		os.Exit(1)
	}
}

// levelNames returns the canonical, upper-case name for a level. It is
// used by the JSON encoder and is the single source of truth for the
// "level" field on the wire.
func levelNames(l Level) string { return l.String() }

// =============================================================================
// Level methods
// =============================================================================

// Debug logs a message at DEBUG level.
func (l *Logger) Debug(msg string) { l.log(LevelDebug, msg) }
func (l *Logger) Debugf(f string, a ...any) {
	l.log(LevelDebug, fmt.Sprintf(f, a...))
}

// Info logs a message at INFO level.
func (l *Logger) Info(msg string) { l.log(LevelInfo, msg) }
func (l *Logger) Infof(f string, a ...any) {
	l.log(LevelInfo, fmt.Sprintf(f, a...))
}

// Warn logs a message at WARN level.
func (l *Logger) Warn(msg string) { l.log(LevelWarn, msg) }
func (l *Logger) Warnf(f string, a ...any) {
	l.log(LevelWarn, fmt.Sprintf(f, a...))
}

// Error logs a message at ERROR level.
func (l *Logger) Error(msg string) { l.log(LevelError, msg) }
func (l *Logger) Errorf(f string, a ...any) {
	l.log(LevelError, fmt.Sprintf(f, a...))
}

// Fatal logs a message at FATAL level then calls os.Exit(1).
func (l *Logger) Fatal(msg string) { l.log(LevelFatal, msg) }
func (l *Logger) Fatalf(f string, a ...any) {
	l.log(LevelFatal, fmt.Sprintf(f, a...))
}
