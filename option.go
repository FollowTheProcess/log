package log

import "time"

// Option is a functional option for configuring a [Logger].
type Option func(*Logger)

// WithLevel sets the log level, that is; the minimum level of logs that will show up.
func WithLevel(level Level) Option {
	return func(l *Logger) {
		l.level = level
	}
}

// TimeFormat sets the format of the time information.
//
// The layout is the standard Go [time.Format] and defaults to [time.RFC3339].
func TimeFormat(format string) Option {
	return func(l *Logger) {
		l.timeFormat = format
	}
}

// TimeFunc sets the mechanism by which the logger knows the current time.
//
// Most usage will not set this option, but it's handy if you want to provide
// a deterministic time for your logs such as during testing etc.
//
// The [Logger] will default to [time.Now] (with UTC) if this option is not set.
func TimeFunc(fn func() time.Time) Option {
	return func(l *Logger) {
		l.timeFunc = fn
	}
}
