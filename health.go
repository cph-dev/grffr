package grffr

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Healthcheck defines the interface for health checks.
//
// Any component or service that can report on own health need to implement
// this interface.
//
// Notice that the only mandatory field is the Status.
type Healthchecker interface {
	Healthcheck() Health
}

type Health struct {
	Status    HealthStatus   `json:"status"`
	Uptime    string         `json:"uptime,omitempty"`
	UptimeSec int64          `json:"uptime_sec,omitempty"`
	Details   map[string]any `json:"details,omitempty"`
	Meta      HealthMeta     `json:"meta,omitzero"`
}

type HealthMeta struct {
	Timestamp        time.Time `json:"timestamp,omitzero"`
	TimestampUnixMs  int64     `json:"timestamp_unix_ms,omitzero"`
	TimestampUnixSec int64     `json:"timestamp_unix_sec,omitzero"`
	Version          string    `json:"version,omitempty"`
}

type HealthStatus string

const (
	HealthStatusOK       = "OK"
	HealthStatusUp       = "UP"
	HealthStatusDown     = "DOWN"
	HealthStatusDegraded = "DEGRADED"
)

func (a *App) defaultLivenessHandler() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if a.isShuttingDown.Load() {
			http.Error(w, "Shutting down", http.StatusServiceUnavailable)
			return
		}
		fmt.Fprintln(w, HealthStatusOK)
	}
	return http.HandlerFunc(fn)
}

func (a *App) defaultStartupHandler() http.Handler {
	return a.defaultReadinessHandler()
}

func (a *App) defaultReadinessHandler() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if a.isShuttingDown.Load() {
			http.Error(w, "Shutting down", http.StatusServiceUnavailable)
			return
		}
		fmt.Fprintln(w, HealthStatusUp)
	}
	return http.HandlerFunc(fn)
}

func (a *App) defaultStatusHandler() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if a.isShuttingDown.Load() {
			http.Error(w, "Shutting down", http.StatusServiceUnavailable)
			return
		}

		now := time.Now()
		uptime := time.Since(a.startedAt)
		uptime = uptime.Round(time.Second)
		health := Health{
			Status:    HealthStatusUp,
			Uptime:    uptime.String(),
			UptimeSec: int64(uptime.Seconds()),
			Details:   map[string]any{},
			Meta: HealthMeta{
				Timestamp:        now.Truncate(time.Millisecond),
				TimestampUnixMs:  now.UnixMilli(),
				TimestampUnixSec: now.Unix(),
				// TODO: Add Version
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(health)
	}
	return http.HandlerFunc(fn)
}
