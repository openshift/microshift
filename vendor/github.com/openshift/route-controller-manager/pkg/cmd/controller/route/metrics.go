package route

import (
	"context"
)

// AuthenticatorMetrics specifies a set of methods that are used to register various metrics
type AuthenticatorMetrics struct {
	// RecordRequestTotal increments the total number of requests for webhooks
	RecordRequestTotal func(ctx context.Context, code string)

	// RecordRequestLatency measures request latency in seconds for webhooks. Broken down by status code.
	RecordRequestLatency func(ctx context.Context, code string, latency float64)
}

type noopMetrics struct{}

func (noopMetrics) RequestTotal(context.Context, string)            {}
func (noopMetrics) RequestLatency(context.Context, string, float64) {}
