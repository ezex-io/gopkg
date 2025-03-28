package logger

type Logger interface {
	// Debug logs a message at debug level.
	// Use for verbose internal events like tracing or dev info.
	Debug(msg string, args ...any)

	// Info logs a message at info level.
	// Use for general application events such as startup or successful actions.
	Info(msg string, args ...any)

	// Warn logs a message at warning level.
	// Use when something unexpected happened but the app can continue.
	Warn(msg string, args ...any)

	// Error logs a message at error level.
	// Use for runtime or business errors that need investigation.
	Error(msg string, args ...any)

	// Fatal logs a message at error level and exits the application with status 1.
	// Use for unrecoverable conditions (e.g., failed to start, config missing).
	Fatal(msg string, args ...any)

	// With returns a new Logger instance with additional context fields.
	// Use to attach fields like "module", "request_id", or "user_id".
	With(args ...any) Logger
}
