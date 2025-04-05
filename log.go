// Package log provides a Logger primarily intended for terminal based command line programs.
//
// The logs it produces are semi-structured with key value pairs being formatted as key=value but are primarily
// intended to be human readable and easy on the eye with a good choice of colours, ideal for command line
// applications that have a --debug or --verbose flag that enables extra logging.
//
// log emphasis simplicity and efficiency so there aren't too many knobs to twiddle, you just get a consistent,
// easy to use, simple logger with minimal overhead.
package log

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/FollowTheProcess/hue"
)

// Styles.
const (
	timestampStyle = hue.Dim
	prefixStyle    = hue.Dim | hue.Bold
	keyStyle       = hue.Magenta
	debugStyle     = hue.Blue | hue.Bold
	infoStyle      = hue.Cyan | hue.Bold
	warnStyle      = hue.Yellow | hue.Bold
	errorStyle     = hue.Red | hue.Bold
)

// missingValue is the placeholder text for a missing value in a log line's key value pair.
const missingValue = "<MISSING>"

// ctxKey is the unexported type used for context key so this key never collides with another.
type ctxKey struct{}

// contextKey is the actual key used to store and retrieve a Logger from a Context.
var contextKey = ctxKey{}

// Logger is a command line logger. It is safe to use across concurrently
// executing goroutines.
type Logger struct {
	w          io.Writer        // Where to write logs to
	timeFunc   func() time.Time // A function to get the current time, defaults to [time.Now] (with UTC)
	mu         *sync.Mutex      // Protects w
	timeFormat string           // The time format layout string, defaults to [time.RFC3339]
	prefix     string           // Optional prefix to prepend to all log messages
	kv         []any            // Persistent key value pairs
	level      Level            // The configured level of this logger, logs below this level are not shown
	isDiscard  atomic.Bool      // w == [io.Discard], cached
}

// New returns a new [Logger] configured to write to w.
//
// The logger can be configured by passing a number of functional options to set
// things like level, prefix etc.
func New(w io.Writer, options ...Option) *Logger {
	logger := &Logger{
		w:          w,
		level:      LevelInfo,
		timeFormat: time.RFC3339,
		timeFunc:   func() time.Time { return time.Now().UTC() },
		mu:         &sync.Mutex{},
	}

	logger.isDiscard.Store(w == io.Discard)

	for _, option := range options {
		option(logger)
	}

	return logger
}

// WithContext stores the given logger in a [context.Context].
//
// The logger may be retrieved from the context with [FromContext].
func WithContext(ctx context.Context, logger *Logger) context.Context {
	return context.WithValue(ctx, contextKey, logger)
}

// FromContext returns the [Logger] from a [context.Context].
//
// If the context does not contain a logger, a default logger is returned.
func FromContext(ctx context.Context) *Logger {
	logger, ok := ctx.Value(contextKey).(*Logger)
	if !ok || logger == nil {
		logger = New(os.Stderr)
	}

	return logger
}

// With returns a new [Logger] with the given persistent key value pairs.
//
// The returned logger is otherwise an exact clone of the caller.
func (l *Logger) With(kv ...any) *Logger {
	sub := l.clone()

	sub.kv = slices.Concat(sub.kv, kv)
	return sub
}

// Prefixed returns a new [Logger] with the given prefix.
//
// The returned logger is otherwise an exact clone of the caller.
func (l *Logger) Prefixed(prefix string) *Logger {
	sub := l.clone()

	sub.prefix = prefix
	return sub
}

// Debug writes a debug level log line.
func (l *Logger) Debug(msg string, kv ...any) {
	l.log(LevelDebug, msg, kv...)
}

// Info writes an info level log line.
func (l *Logger) Info(msg string, kv ...any) {
	l.log(LevelInfo, msg, kv...)
}

// Warn writes a warning level log line.
func (l *Logger) Warn(msg string, kv ...any) {
	l.log(LevelWarn, msg, kv...)
}

// Error writes an error level log line.
func (l *Logger) Error(msg string, kv ...any) {
	l.log(LevelError, msg, kv...)
}

// log logs the given levelled message.
func (l *Logger) log(level Level, msg string, kv ...any) {
	if l.isDiscard.Load() || l.level > level {
		// Do as little work as possible
		return
	}

	// Buffer the output as e.g. stderr is not buffered by default. Do this
	// by fetching and putting buffers from a [sync.Pool] so we don't have to
	// constantly allocate new buffers
	buf := getBuffer()
	defer putBuffer(buf)

	buf.WriteString(timestampStyle.Text(l.timeFunc().Format(l.timeFormat)))
	buf.WriteByte(' ')
	buf.WriteString(level.styled())
	if l.prefix != "" {
		buf.WriteString(" " + prefixStyle.Text(l.prefix))
	}
	buf.WriteByte(':')
	buf.WriteByte(' ')
	buf.WriteString(msg)

	if numKVs := len(l.kv) + len(kv); numKVs != 0 {
		kvs := make([]any, 0, numKVs)
		kvs = append(kvs, l.kv...)
		if len(kvs)%2 != 0 {
			kvs = append(kvs, missingValue)
		}

		kvs = append(kvs, kv...)
		if len(kvs)%2 != 0 {
			kvs = append(kvs, missingValue)
		}

		for i := 0; i < len(kvs); i += 2 {
			buf.WriteByte(' ')
			key := keyStyle.Sprint(kvs[i])
			val := fmt.Sprintf("%+v", kvs[i+1])

			if needsQuotes(val) || val == "" {
				val = strconv.Quote(val)
			}

			buf.WriteString(key)
			buf.WriteByte('=')
			buf.WriteString(val)
		}
	}

	buf.WriteByte('\n')

	// WriteTo drains the buffer
	l.mu.Lock()
	defer l.mu.Unlock()
	buf.WriteTo(l.w) //nolint: errcheck // Just like printing
}

// clone returns an exact clone of the calling logger.
func (l *Logger) clone() *Logger {
	clone := &Logger{
		w:          l.w,
		timeFunc:   l.timeFunc,
		timeFormat: l.timeFormat,
		prefix:     l.prefix,
		level:      l.level,
		mu:         l.mu,
	}

	clone.isDiscard.Store(l.isDiscard.Load())

	return clone
}

// Each log method (Debug, Info, Warn) etc. gets a buffer from this pool
// so as not to keep re-allocating and destroying them.
var bufPool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

// getBuffer fetches a buffer from the pool, the returned buffer
// is empty and ready to use.
func getBuffer() *bytes.Buffer {
	buf := bufPool.Get().(*bytes.Buffer) //nolint:revive,errcheck,forcetypeassert // We are in total control of this
	buf.Reset()
	return buf
}

// putBuffer puts the buffer back into the pool.
func putBuffer(buf *bytes.Buffer) {
	// Proper usage of a sync.Pool requires each entry to have approximately
	// the same memory cost. To obtain this property when the stored type
	// contains a variably-sized buffer, we add a hard limit on the maximum buffer
	// to place back in the pool.
	//
	// See https://go.dev/issue/23199

	// Approx 65kb
	const maxSize = 64 << 10
	if buf.Cap() > maxSize {
		// Make the buffer nil so GC cleans it up
		buf = nil
	}

	bufPool.Put(buf)
}

// needsQuotes returns whether s should be displayed as "s".
func needsQuotes(s string) bool {
	for _, char := range s {
		if char == utf8.RuneError || unicode.IsSpace(char) || !unicode.IsPrint(char) {
			return true
		}
	}

	return false
}
