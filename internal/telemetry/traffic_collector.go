// Package telemetry agrège de manière atomique les compteurs de trafic et de latence.
package telemetry

import (
	"sync/atomic"
	"time"
)

// MetricsSnapshot représente une vue cohérente et sérialisable en JSON de l'état du proxy.
type MetricsSnapshot struct {
	TotalRequests     uint64  `json:"total_requests"`
	AllowedRequests   uint64  `json:"allowed_requests"`
	RejectedRateLimit uint64  `json:"rejected_429"`
	UpstreamErrors    uint64  `json:"upstream_5xx"`
	RejectionRatioPct float64 `json:"rejection_ratio_pct"`
	MeanLatencyMs     float64 `json:"mean_latency_ms"`
}

// TrafficCollector enregistre les événements HTTP sans verrou bloquant (atomics 64 bits).
type TrafficCollector struct {
	totalRequests   atomic.Uint64
	allowedRequests atomic.Uint64
	rejected429     atomic.Uint64
	upstreamErrors  atomic.Uint64
	totalLatencyUs  atomic.Uint64
}

// RecordAllowed comptabilise une requête relayée avec sa durée de bout en bout.
func (c *TrafficCollector) RecordAllowed(statusCode int, elapsed time.Duration) {
	c.totalRequests.Add(1)
	c.allowedRequests.Add(1)
	c.totalLatencyUs.Add(uint64(max(0, elapsed.Microseconds())))
	if statusCode >= 500 {
		c.upstreamErrors.Add(1)
	}
}

// RecordRateLimited comptabilise un rejet défensif immédiat (HTTP 429).
func (c *TrafficCollector) RecordRateLimited() {
	c.totalRequests.Add(1)
	c.rejected429.Add(1)
}

// Snapshot calcule une photographie pure des indicateurs clés à l'instant courant.
func (c *TrafficCollector) Snapshot() MetricsSnapshot {
	total := c.totalRequests.Load()
	allowed := c.allowedRequests.Load()
	rejected := c.rejected429.Load()
	upErr := c.upstreamErrors.Load()
	latUs := c.totalLatencyUs.Load()

	var rejectPct float64
	if total > 0 {
		rejectPct = float64(rejected) * 100.0 / float64(total)
	}

	var meanMs float64
	if allowed > 0 {
		meanMs = (float64(latUs) / float64(allowed)) / 1000.0
	}

	return MetricsSnapshot{
		TotalRequests:     total,
		AllowedRequests:   allowed,
		RejectedRateLimit: rejected,
		UpstreamErrors:    upErr,
		RejectionRatioPct: rejectPct,
		MeanLatencyMs:     meanMs,
	}
}
