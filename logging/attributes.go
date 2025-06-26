package logging

import "log/slog"

// Error returns an slog attribute with the error value.
func Error(err error) slog.Attr {
	return slog.Any("error", err)
}
