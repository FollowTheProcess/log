package log

// Level is a log level.
type Level int

const (
	// LevelDebug is the debug log level, intended for verbose logging modes or internal debugging.
	LevelDebug Level = -4

	// LevelInfo is the info log level, this is the default level, intended for progress updates
	// and other informational messages.
	LevelInfo Level = 0

	// LevelWarn is the warning log level, intended for raising recoverable issues to the user. Warning
	// logs should refer to events that are worth flagging to the user but the program can easily
	// recover from such as a missing configuration file when the application can fall back to defaults.
	LevelWarn Level = 4

	// LevelError is the error log level. This is the highest log level provided by log and is intended
	// for signalling non-recoverable errors to the user. Typically followed by an actual go error and
	// possibly program exit.
	LevelError Level = 8
)

const (
	debugString = "DEBUG"
	infoString  = "INFO"
	warnString  = "WARN"
	errorString = "ERROR"
)

// String returns the stylised representation of the log level.
func (l Level) String() string {
	switch l {
	case LevelDebug:
		return debugStyle.Text(debugString)
	case LevelInfo:
		return infoStyle.Text(infoString)
	case LevelWarn:
		return warnStyle.Text(warnString)
	case LevelError:
		return errorStyle.Text(errorString)
	default:
		return "unknown"
	}
}
