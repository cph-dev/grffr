package grffr

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

func (a *App) initWebServer() error {
	port, err := strconv.Atoi(os.Getenv("HTTP_PORT"))
	if err != nil {
		slog.Debug("Using default port :80")
		port = 80
	}

	err = portAvailable(port)
	if err != nil {
		return fmt.Errorf("port %d unavailable: %w", port, err)
	}

	mux := chi.NewMux()

	// Health checks, propes and status
	mux.Route("/.well-known/health", func(r chi.Router) {

		// Is application started up?
		//
		// Don't do other health checks until the application has
		// declared itself as started.
		if a.configuration.StartupHandler != nil {
			r.Handle("GET /startup", a.configuration.StartupHandler)
		}

		// Is application alive?
		//
		// If not alive then the application/process/container should be
		// terminated and not be sent any traffic.
		if a.configuration.LivenessHandler != nil {
			r.Handle("GET /alive", a.configuration.LivenessHandler)
		}

		// Is application ready to serve?
		//
		// Determines if the application is currently ready to serve traffic.
		// If not then wait some time and ask again if ready.
		// Not being ready is a temporary state unlike not being alive.
		if a.configuration.ReadinessHandler != nil {
			r.Handle("GET /ready", a.configuration.ReadinessHandler)
		}

		// What's the health status of the application?
		//
		// This is a more detailed report on the applications health.
		if a.configuration.StatusHandler != nil {
			r.Handle("GET /status", a.configuration.StatusHandler)
		}
	})

	// Configure web server
	inflightCtx, inflightCancel := context.WithCancel(context.Background())
	a.httpServer = http.Server{
		Addr:    fmt.Sprintf(":%v", port),
		Handler: mux,

		// TODO: Timeouts should probably not be hard-coded in a framework like this
		// as use cases are unknown and some services might need short timeouts, and other might
		//  need longer timeouts.
		ReadTimeout:       2 * time.Minute,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      time.Minute,
		IdleTimeout:       time.Minute,
		BaseContext: func(net.Listener) context.Context {
			return inflightCtx
		},
	}

	// TODO: This is probably not correct.
	// When should inflight context be cancelled?
	a.httpServer.RegisterOnShutdown(func() {
		slog.Debug("Cancel inflight context")
		inflightCancel()
	})

	return nil
}
