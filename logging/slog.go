package logging

import (
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
)

// Configure sets up the logger based on the environment.
//
// Production gets JSON output, while development gets pretty colored output.
func Configure() *slog.Logger {
	// TODO: This should be based on IsTTY
	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}

	var logger *slog.Logger
	if env == "development" || env == "dev" {
		tinted := tint.NewHandler(os.Stdout, &tint.Options{
			AddSource:  true,
			Level:      slog.LevelDebug,
			TimeFormat: time.TimeOnly,
		})
		handler := &ContextHandler{
			Handler: tinted,
		}
		logger = slog.New(handler)
		logger.Debug("Amazing logging configured.",
			slog.String("have", "fun"),
			slog.String("be", "awesome"),
			slog.String("drink", "water"),
		)
	} else {
		loggerOpts := &slog.HandlerOptions{
			AddSource: true,
		}
		jsonHandler := slog.NewJSONHandler(os.Stdout, loggerOpts)
		h := &ContextHandler{
			Handler: jsonHandler,
		}
		logger = slog.New(h).With(
			slog.Group("app",
				slog.String("env", env),
				// slog.String("version", version),
			))
	}

	return logger
}
