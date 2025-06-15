package options

import "log/slog"

// WithLogger sets the logger to use within the framework.
func WithLogger(logger *slog.Logger) Option {
	return func(cfg *Configuration) {
		cfg.Logger = logger
	}
}
