package options

import (
	"log/slog"
	"net/http"
)

// TODO: The health checks should also come with configuration for path
// so that can easilly be overriden with an Option.
//
// Right now an overriden health check will be mounted on `/.well-known/xx`
// which might not fit into an organisation's standards.

// Option can be applied to Configuration when calling New to change defaults.
type Option func(*Configuration)

type Configuration struct {
	Debug bool

	// Logger to use in application.
	Logger *slog.Logger

	// Purpose: To indicate whether the container
	// is running. If the liveness probe fails, the
	// container will be restarted.
	//
	// Default location is /.well-known/health/alive
	LivenessHandler http.Handler

	// Purpose: To indicate whether the container
	// is ready to serve traffic. If the readiness
	// probe fails, the container will be removed
	// from the pool of available containers.
	//
	// Default location is /.well-known/health/ready
	ReadinessHandler http.Handler

	// Purpose: To indicate whether the application
	// has started successfully. This probe is useful
	// for applications that take a long time to start.
	//
	// Default is to share the readiness handler and mounted
	// on /.well-known/health/startup
	StartupHandler http.Handler

	// Purpose: To give detailed insight into the health of
	// the application included dependencies.
	//
	// Default location is /.well-known/health/status
	StatusHandler http.Handler
}
