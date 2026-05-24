// Package log provides a Logger primarily intended for terminal based command line programs.
//
// The logs it produces are semi-structured with key value pairs being formatted as key=value but are primarily
// intended to be human readable and easy on the eye with a good choice of colours, ideal for command line
// applications that have a --debug or --verbose flag that enables extra logging.
//
// log emphasises simplicity and efficiency so there aren't too many knobs to twiddle, you just get a consistent,
// easy to use, simple logger with minimal overhead.
package log // import "go.followtheprocess.codes/log"

import (
	"io"
	"log/slog"
	"slices"
	"strconv"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"

	"go.followtheprocess.codes/hue"
)

const (
	// scratchSize is the size of the stack buffer used to format a timestamp
	// before styling it; comfortably larger than any standard time layout.
	scratchSize = 64

	// bufferSize is the initial capacity of a pooled log line buffer, chosen to
	// hold a typical line without reallocating.
	bufferSize = 256

	// base10 is the radix used to format integer attribute values.
	base10 = 10

	// float64Bits is the bit size used to format floating point attribute values.
	float64Bits = 64
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

// Logger is a command line logger. It is safe to use across concurrently
// executing goroutines.
//
// The zero value is not usable; construct a Logger with [New].
type Logger struct {
	w          io.Writer        // Where to write logs to
	timeFunc   func() time.Time // A function to get the current time, defaults to [time.Now] (with UTC)
	mu         *sync.Mutex      // Protects w, pointer so that child loggers share the same mutex
	timeFormat string           // The time format layout string, defaults to [time.RFC3339]
	prefix     []byte           // Optional prefix to prepend to all log messages, stored as bytes for the hot path
	attrs      []slog.Attr      // Persistent key value pairs
	level      Level            // The configured level of this logger, logs below this level are not shown
	isDiscard  bool             // w == [io.Discard], cached. Only written during construction, before the logger is shared
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
		isDiscard:  w == io.Discard,
	}

	for _, option := range options {
		option(logger)
	}

	return logger
}

// With returns a new [Logger] with the given persistent key value pairs.
//
// The returned logger is otherwise an exact clone of the caller.
func (l *Logger) With(attrs ...slog.Attr) *Logger {
	sub := l.clone()

	sub.attrs = slices.Concat(sub.attrs, attrs)

	return sub
}

// Prefixed returns a new [Logger] with the given prefix.
//
// The returned logger is otherwise an exact clone of the caller.
func (l *Logger) Prefixed(prefix string) *Logger {
	sub := l.clone()

	sub.prefix = []byte(prefix)

	return sub
}

// Debug writes a debug level log line.
func (l *Logger) Debug(msg string, attrs ...slog.Attr) {
	l.log(LevelDebug, msg, attrs...)
}

// Info writes an info level log line.
func (l *Logger) Info(msg string, attrs ...slog.Attr) {
	l.log(LevelInfo, msg, attrs...)
}

// Warn writes a warning level log line.
func (l *Logger) Warn(msg string, attrs ...slog.Attr) {
	l.log(LevelWarn, msg, attrs...)
}

// Error writes an error level log line.
func (l *Logger) Error(msg string, attrs ...slog.Attr) {
	l.log(LevelError, msg, attrs...)
}

// log logs the given levelled message.
func (l *Logger) log(level Level, msg string, attrs ...slog.Attr) {
	if l.isDiscard || l.level > level {
		// Do as little work as possible
		return
	}

	// Build the line in a byte buffer fetched from a [sync.Pool] so we don't
	// constantly allocate. Styled, known-ahead text (timestamp, level, prefix)
	// is appended with hue's allocation-free AppendText.
	bufp := getBuffer()
	defer putBuffer(bufp)

	// Dereference the working copy so we don't have to dereference every call
	buf := *bufp

	// Format the timestamp into a stack scratch buffer so we avoid allocating
	// an intermediate string before styling it.
	var scratch [scratchSize]byte

	timestamp := l.timeFunc().AppendFormat(scratch[:0], l.timeFormat)
	buf = timestampStyle.AppendText(buf, timestamp)

	buf = append(buf, ' ')
	buf = level.appendTo(buf)

	if len(l.prefix) != 0 {
		buf = append(buf, ' ')
		buf = prefixStyle.AppendText(buf, l.prefix)
	}

	buf = append(buf, ':')

	// DEBUG and ERROR are 5 characters, INFO and WARN are 4. Pad the shorter
	// labels with an extra space so the message always starts in the same column.
	buf = append(buf, ' ')
	if level == LevelInfo || level == LevelWarn {
		buf = append(buf, ' ')
	}

	buf = append(buf, msg...)

	for _, attr := range l.attrs {
		buf = appendAttr(buf, attr)
	}

	for _, attr := range attrs {
		buf = appendAttr(buf, attr)
	}

	buf = append(buf, '\n')

	// Put it back
	*bufp = buf

	l.mu.Lock()
	defer l.mu.Unlock()

	l.w.Write(buf) //nolint: errcheck // Just like printing
}

// appendAttr appends a single " key=value" pair to dst and returns the
// extended slice. The key is quoted if it contains whitespace or is empty.
func appendAttr(dst []byte, attr slog.Attr) []byte {
	dst = append(dst, ' ')

	key := attr.Key
	if key == "" || needsQuotes(key) {
		key = strconv.Quote(key)
	}

	dst = keyStyle.AppendString(dst, key)
	dst = append(dst, '=')

	return appendValue(dst, attr.Value)
}

// appendValue appends the textual form of v to dst and returns the extended slice.
//
// Scalar kinds are written straight into the buffer, skipping the "needs quotes" check
// as their text can never contain whitespace and so are never quoted.
//
// Other kinds fall back to [slog.Value.String], quoted if they contain whitespace or are empty.
func appendValue(dst []byte, v slog.Value) []byte {
	// Resolve any [slog.LogValuer]
	// See https://github.com/golang/example/blob/master/slog-handler-guide/README.md
	if v.Kind() == slog.KindLogValuer {
		v = v.Resolve()
	}

	switch v.Kind() {
	case slog.KindInt64:
		return strconv.AppendInt(dst, v.Int64(), base10)
	case slog.KindUint64:
		return strconv.AppendUint(dst, v.Uint64(), base10)
	case slog.KindFloat64:
		return strconv.AppendFloat(dst, v.Float64(), 'g', -1, float64Bits)
	case slog.KindBool:
		return strconv.AppendBool(dst, v.Bool())
	default:
		s := v.String()
		if s == "" || needsQuotes(s) {
			return strconv.AppendQuote(dst, s)
		}

		return append(dst, s...)
	}
}

// clone returns an exact clone of the calling logger.
func (l *Logger) clone() *Logger {
	clone := &Logger{
		w:          l.w,
		timeFunc:   l.timeFunc,
		timeFormat: l.timeFormat,
		prefix:     l.prefix,
		attrs:      l.attrs,
		level:      l.level,
		mu:         l.mu,
		isDiscard:  l.isDiscard,
	}

	return clone
}

// Each log method (Debug, Info, Warn) etc. gets a buffer from this pool
// so as not to keep re-allocating and destroying them.
var bufPool = sync.Pool{
	New: func() any {
		buf := make([]byte, 0, bufferSize)
		return &buf
	},
}

// getBuffer fetches a buffer from the pool, the returned buffer
// is empty and ready to use.
func getBuffer() *[]byte {
	bufp := bufPool.Get().(*[]byte) //nolint:revive,errcheck,forcetypeassert // We are in total control of this
	*bufp = (*bufp)[:0]             // Reset

	return bufp
}

// putBuffer puts the buffer back into the pool.
func putBuffer(bufp *[]byte) {
	// Proper usage of a sync.Pool requires each entry to have approximately
	// the same memory cost. To obtain this property when the stored type
	// contains a variably-sized buffer, we add a hard limit on the maximum buffer
	// to place back in the pool.
	//
	// See https://go.dev/issue/23199

	// Approx 65kb
	const maxSize = 64 << 10
	if cap(*bufp) > maxSize {
		return
	}

	bufPool.Put(bufp)
}

// needsQuotes returns whether s should be displayed as "s".
func needsQuotes(s string) bool {
	for i := 0; i < len(s); {
		// ASCII fast path: most keys and values are printable ASCII
		// Anything <= space (control characters and space itself) or DEL needs quoting.
		if b := s[i]; b < utf8.RuneSelf {
			if b <= ' ' || b == 0x7f {
				return true
			}

			i++

			continue
		}

		char, size := utf8.DecodeRuneInString(s[i:])
		if char == utf8.RuneError || unicode.IsSpace(char) || !unicode.IsPrint(char) {
			return true
		}

		i += size
	}

	return false
}
