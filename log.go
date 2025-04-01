// Package log provides a Logger primarily intended for terminal based command line programs.
//
// The logs it produces are semi-structured with key value pairs being formatted as key=value but are primarily
// intended to be human readable and easy on the eye with a good choice of colours, ideal for command line
// applications that have a --debug or --verbose flag that enables extra logging.
package log

import (
	"bytes"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/FollowTheProcess/hue"
)

// Styles.
const (
	timestampStyle = hue.BrightBlack
	debugStyle     = hue.Blue | hue.Bold
	infoStyle      = hue.Cyan | hue.Bold
	warnStyle      = hue.Yellow | hue.Bold
	errorStyle     = hue.Red | hue.Bold
)

// Logger is a command line logger. It is safe to use across concurrently
// executing goroutines.
type Logger struct {
	w          io.Writer        // Logs are written here, if it is [io.Discard], log methods exit early having done nothing
	timeFunc   func() time.Time // Function called to get the current time, defaults to [time.Now] (UTC)
	timeFormat string           // The time serialisation format, defaults to [time.RFC3339]
	level      Level            // The level at which this logger is set
	mu         sync.Mutex       // Protects writing to w
	isDiscard  atomic.Bool      // w == io.Discard, cached as can only be set once via [New]
}

// New returns a new [Logger].
func New(w io.Writer, options ...Option) *Logger {
	logger := &Logger{
		w:          w,
		level:      LevelInfo,
		timeFormat: time.RFC3339,
		timeFunc:   now,
	}

	logger.isDiscard.Store(w == io.Discard)

	for _, option := range options {
		option(logger)
	}

	return logger
}

// Debug writes a debug level log line.
func (l *Logger) Debug(msg string) {
	l.log(LevelDebug, msg)
}

// Info writes an info level log line.
func (l *Logger) Info(msg string) {
	l.log(LevelInfo, msg)
}

// Warn writes a warning level log line.
func (l *Logger) Warn(msg string) {
	l.log(LevelWarn, msg)
}

// Error writes an error level log line.
func (l *Logger) Error(msg string) {
	l.log(LevelError, msg)
}

// log logs the given levelled message.
func (l *Logger) log(level Level, msg string) {
	if l.isDiscard.Load() || l.level > level {
		// Do as little work as possible
		return
	}

	// Buffer the output as e.g. stderr is not buffered by default. Do this
	// by fetching and putting buffers from a [sync.Pool] so we don't have to
	// constantly allocate new buffers
	buf := getBuffer()
	defer putBuffer(buf)

	fmt.Fprintf(
		buf,
		"%s %s: %s\n",
		timestampStyle.Text(time.Now().Format(l.timeFormat)),
		level.styled(),
		msg,
	)

	// WriteTo drains the buffer
	l.mu.Lock()
	defer l.mu.Unlock()
	buf.WriteTo(l.w) //nolint: errcheck // Just like printing
}

// Each log method (Debug, Info, Warn) etc.
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

// now returns the current time with UTC.
func now() time.Time {
	return time.Now().UTC()
}
