package options

import "net/http"

// DisableDefaultHealthHandlers prevents the following default
// handlers to be registered during initialisation:
// Default LivenessHandler, ReadinessHandler, StartupHandler
// and StatusHandler.
//
// You can register your own checks either as a handler like any other
// end-point, or you can provide checks.
func DisableDefaultHealthHandlers(cfg *Configuration) {
	cfg.StartupHandler = nil
	cfg.LivenessHandler = nil
	cfg.ReadinessHandler = nil
	cfg.StatusHandler = nil
}

// WithHealthStartupCheck overrides the default start up check with
// the given handler.
//
// The end-point is still served on the default location of
//
//	GET /.well-known/health/startup
func WithHealthStartupCheck(h http.Handler) Option {
	return func(cfg *Configuration) {
		cfg.LivenessHandler = h
	}
}

// WithHealthLivenessCheck overrides the default liveness health check with
// the given handler.
//
// The end-point is still served on the default location of
//
//	GET /.well-known/health/alive
func WithHealthLivenessCheck(h http.Handler) Option {
	return func(cfg *Configuration) {
		cfg.LivenessHandler = h
	}
}

// WithHealthReadinessCheck overrides the default readiness health check with
// the given handler.
//
// The end-point is still served on the default location of
//
//	GET /.well-known/health/readinessHandler
func WithHealthReadinessCheck(h http.Handler) Option {
	return func(cfg *Configuration) {
		cfg.ReadinessHandler = h
	}
}

// WithHealthStatusCheck overrides the default status health check with
// the given handler.
//
// The end-point is still served on the default location of
//
//	GET /.well-known/health/status
func WithHealthStatusCheck(h http.Handler) Option {
	return func(cfg *Configuration) {
		cfg.StatusHandler = h
	}
}
