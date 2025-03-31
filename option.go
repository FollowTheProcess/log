package log

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
