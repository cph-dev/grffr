package grffr

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"sync"

	"github.com/hashicorp/go-multierror"
	"go.cph.dev/grffr/data"
	"go.cph.dev/grffr/logging"
	"go.opentelemetry.io/otel/trace"
)

// AddComponent to application.
//
// The app will start up the component and make sure it is
// also stopped again during the shutdown sequence after
// all incoming requests are drained.cccccbrfgkcdduculkblceldlhevurnlcgbtgiffgvtj
//
// Implement [Name() string] (NamedComponent interface)
// for more context during logging, etc.
func (a *App) AddComponent(c Component) {
	a.components = append(a.components, c)
}

// Component running inside application.
//
// It has a simple life-cycle:
//   - Init() the component before it is started.
//   - Start() starts the component.
//   - Stop() stops the component.
//   - Provided context.Context should be checked
//     for Done state and exit early if needed.
type Component interface {
	Init(ctx context.Context) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

type NamedComponent interface {
	Component
	Named
}

// NamedComponent is a component with a name.
//
// Name will be used to easier identify component in logging,
// tracing, etc.
type Named interface {
	Name() string
}

// WantLogger is a component with a logger.
//
// UseLogger will be called during initialization.
type WantLogger interface {
	UseLogger(*slog.Logger)
}

// WantTracer is a component with a tracer.
//
// UseTracer will be called during initialization.
type WantTracer interface {
	UseTracer(trace.Tracer)
}

// WantSQL is a component with a SQL connection.
//
// UseSQL will be called during initialization.
type WantSQL interface {
	UseSQL(data.SQL)
}

func (a *App) initComponents(ctx context.Context) error {
	var result error
	for c := range slices.Values(a.components) {
		if logger, ok := c.(WantLogger); ok {
			logger.UseLogger(a.logger)
		}
		if tracer, ok := c.(WantTracer); ok {
			tracer.UseTracer(a.tracer)
		}
		if sql, ok := c.(WantSQL); ok {
			sql.UseSQL(a.sql)
		}
		if err := c.Init(ctx); err != nil {
			result = multierror.Append(result, err)
		}
	}

	return result
}

func (a *App) startComponents(
	ctx context.Context,
	exit *sync.WaitGroup,
) {
	for c := range slices.Values(a.components) {
		startCtx := ctx
		if named, ok := c.(Named); ok {
			startCtx = logging.AppendCtx(startCtx,
				slog.Group("component",
					slog.String("named", named.Name()),
					slog.String("type", fmt.Sprintf("%T", c)),
				),
			)
		} else {
			startCtx = logging.AppendCtx(startCtx,
				slog.Group("component",
					slog.String("type", fmt.Sprintf("%T", c)),
				),
			)
		}
		slog.InfoContext(startCtx, "Starting")

		exit.Add(1)
		go func() {
			defer exit.Done()

			if err := c.Start(startCtx); err != nil {
				slog.WarnContext(startCtx, "Starting component failed", logging.Error(err))
			}
		}()
	}
}

func (a *App) stopComponents(ctx context.Context) error {
	var result error
	a.logger.Info("Stopping components.")
	for c := range slices.Values(a.components) {
		if named, ok := c.(Named); ok {
			ctx = logging.AppendCtx(ctx,
				slog.Group("component",
					slog.String("named", named.Name()),
					slog.String("type", fmt.Sprintf("%T", c)),
				),
			)
		} else {
			ctx = logging.AppendCtx(ctx,
				slog.Group("component",
					slog.String("type", fmt.Sprintf("%T", c)),
				),
			)
		}
		a.logger.InfoContext(ctx, "Stopping component")
		if err := c.Stop(ctx); err != nil {
			result = multierror.Append(result, err)
		}
	}

	return result
}
