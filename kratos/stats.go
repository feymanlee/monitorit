//
// Package prometheus
// @Author: feymanlee@gmail.com
// @Description:
// @File:  prometheus
// @Date: 2023/8/8 20:07
//

package kratos

import (
	"context"
	"fmt"

	prom "github.com/go-kratos/kratos/contrib/metrics/prometheus/v2"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/metrics"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	serverMetricSeconds = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "service",
		Subsystem: "requests",
		Name:      "duration_sec",
		Help:      "server requests duration(sec).",
		Buckets:   []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.250, 0.5, 1},
	}, []string{"kind", "operation"})

	clientMetricRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "service",
		Subsystem: "requests",
		Name:      "code_total",
		Help:      "The total number of processed requests",
	}, []string{"kind", "operation", "code", "reason"})

	panicCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "service",
		Subsystem: "runtime",
		Name:      "panic_total",
		Help:      "Total number of panics",
	}, []string{"kind", "operation", "reason"})
)

func init() {
	prometheus.MustRegister(serverMetricSeconds, clientMetricRequests, panicCounter)
}

func PanicInc(ctx context.Context, err any) {
	var (
		kind      string
		operation string
	)
	if info, ok := transport.FromServerContext(ctx); ok {
		kind = info.Kind().String()
		operation = info.Operation()
	}
	panicCounter.With(prometheus.Labels{
		"kind":      kind,
		"operation": operation,
		"reason":    fmt.Sprintf("%s", err),
	}).Inc()
}

func PrometheusMiddleware() middleware.Middleware {
	return metrics.Server(
		metrics.WithSeconds(prom.NewHistogram(serverMetricSeconds)),
		metrics.WithRequests(prom.NewCounter(clientMetricRequests)),
	)
}
