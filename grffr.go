package grffr

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"go.cph.dev/grffr/logging"
	"go.cph.dev/grffr/options"
)

const (

	// readinessDelay is the duration to wait before checking if the server is ready.
	//
	// It is expected that readiness checks are performed to determine if the application
	// is ready for traffic. The application will wait up til this duration with
	// shutting down to ensure proxies/load balancers have had a chance to switch traffic
	// to a new healthy instance.
	readinessDelay = 5 * time.Second

	// shutdownDuration is the duration to wait for components to shutdown.
	shutdownDuration = 15 * time.Second
)

// New creates a new App with the given options applied.
//
// Defaults are applied before options and can be overrriden or entirely disabled.
func New(opts ...options.Option) *App {
	app := App{}

	logger := logging.Configure()
	slog.SetDefault(logger)

	slog.Debug("Setting defaults.")
	cfg := options.Configuration{
		Logger:           logger,
		StartupHandler:   app.defaultStartupHandler(),
		ReadinessHandler: app.defaultReadinessHandler(),
		LivenessHandler:  app.defaultLivenessHandler(),
		StatusHandler:    app.defaultStatusHandler(),
	}

	slog.Debug("Applying options.")
	for _, opt := range opts {
		opt(&cfg)
	}

	app.configuration = cfg

	return &app
}

// App is the main entry-point for creating a new
// App application.
type App struct {
	debug          bool
	logger         *slog.Logger
	startedAt      time.Time
	configuration  options.Configuration
	isShuttingDown atomic.Bool
	httpServer     http.Server
	components     []Component
}

func (a *App) Run() {
	defer func() {
		if err := recover(); err != nil {
			slog.Error("Panic, terminating program.", "error", err)
			os.Exit(1)
		}
		slog.Info("Shutdown complete. Ktxb.")
	}()

	// TODO: Support for start-up deadline?
	ctx := context.Background()

	a.startedAt = time.Now()

	slog.Debug("Initialising Grffr application...")
	err := a.init(ctx)
	if err != nil {
		slog.Error("Initialising Grffr application", "error", err)
		os.Exit(1)
	}

	slog.Info("Starting Grffr application...")
	err = a.run(ctx)
	if err != nil {
		slog.Error("Running Grffr application", "error", err)
		os.Exit(2)
	}
}

func (a *App) init(ctx context.Context) error {
	a.debug = a.configuration.Debug

	if a.configuration.Logger != nil {
		a.logger = a.configuration.Logger
		slog.SetDefault(a.logger)
	}
	return errors.Join(
		a.initComponents(ctx),
		a.initWebServer(),
	)
}

func (a *App) run(ctx context.Context) error {
	var (
		l sync.Mutex
		e error
	)
	addErr := func(err error) {
		l.Lock()
		defer l.Unlock()
		e = errors.Join(e, err)
	}

	ctx, stop := signal.NotifyContext(ctx,
		// We need to use os.Interrupt to gracefully shutdown on Ctrl+C which is SIGINT
		os.Interrupt,

		// syscall.SIGTERM is the usual signal for termination and the default one (it can be modified) for docker containers,
		// which is also used by kubernetes.
		syscall.SIGTERM,
	)
	defer stop()

	var exit sync.WaitGroup

	// Handle signals to shutdown application
	exit.Add(1)
	go func() {
		defer exit.Done()

		// Block until a signal is received, then initate shutdown
		<-ctx.Done()
		stop()
		a.isShuttingDown.Store(true)
		slog.InfoContext(ctx, "Received shut down signal, shutting down application.")

		// TODO: Wait or enure that readiness has been signalled.
		// Either Sleep(readinessDelay) or make some channel communication to
		// signal that the Readiness end-point has been called and responded.

		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownDuration)
		defer cancel()

		err := a.shutdown(shutdownCtx)
		// // Cancel any inflight requests that might still be processing ?
		// inflightCancel()
		if err != nil {
			addErr(fmt.Errorf("shutting down application: %w", err))
		}
	}()

	// Start components
	startUpCtx, startUpCancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer startUpCancel()

	a.startComponents(startUpCtx, &exit)

	// Start web server
	exit.Add(1)
	go func() {
		defer exit.Done()

		slog.Info("Starting HTTP server...")
		err := a.httpServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			addErr(fmt.Errorf("HTTP server stopped with unexpected error: %w", err))
		}
	}()

	// Wait for all components to exit
	exit.Wait()
	slog.InfoContext(ctx, "Application stopped.")

	return e
}

// shutdown components and services.
//
// Services are shutdown first to ensure request drain, then
// components are stopped.
func (a *App) shutdown(ctx context.Context) error {
	err := errors.Join(
		a.httpServer.Shutdown(ctx),
		a.stopComponents(ctx),
	)
	if err != nil {
		slog.ErrorContext(ctx, "Shutdown", "error", err)
		return err
	}

	return nil
}

// portAvailable checks if the given port is available for use.
//
// Error is nil if available, otherwise error contains the underlying reason.
func portAvailable(port int) error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		return err
	}
	defer l.Close()

	return nil
}
