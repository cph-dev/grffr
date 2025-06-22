package grffr

import (
	"context"
	"slices"
	"sync"

	"github.com/hashicorp/go-multierror"
)

// AddComponent to application.
//
// It will be started when the apps Start() is called.
// And stopped again during the shutdown sequence after
// all incoming requests are drained.
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

type Named interface {
	Name() string
}

func (a *App) initComponents(ctx context.Context) error {
	var result error
	for c := range slices.Values(a.components) {
		if err := c.Init(ctx); err != nil {
			result = multierror.Append(result, err)
		}
	}

	return result
}
func (a *App) startComponents(
	ctx context.Context,
	exit *sync.WaitGroup,
) error {
	var result error

	for c := range slices.Values(a.components) {

		exit.Add(1)
		go func() {
			defer exit.Done()

			if err := c.Start(ctx); err != nil {
				result = multierror.Append(result, err)
			}
		}()
	}

	return result
}

func (a *App) stopComponents(ctx context.Context) error {
	var result error

	for c := range slices.Values(a.components) {
		if err := c.Stop(ctx); err != nil {
			result = multierror.Append(result, err)
		}
	}

	return result
}
