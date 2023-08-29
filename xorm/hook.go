//
// Package xorm
// @Author: feymanlee@gmail.com
// @Description:
// @File:  hook
// @Date: 2023/8/28 19:25
//

package xorm

import (
	"context"
	"strings"

	"github.com/feymanlee/monitorit"
	"github.com/prometheus/client_golang/prometheus"
	"xorm.io/xorm/contexts"
)

type Hook struct {
	options        *Options
	instanceName   string
	queryHistogram *prometheus.HistogramVec
	queryCounter   *prometheus.CounterVec
	errorCounter   *prometheus.CounterVec
}

var (
	queryLabelNames = []string{"db_name", "command"}
	errorLabelNames = []string{"db_name", "command", "error"}
)

func NewHook(dbName string, opts ...Option) *Hook {
	options := DefaultOptions()
	options.Merge(opts...)
	c := Hook{
		options:      options,
		instanceName: dbName,
	}
	c.queryHistogram = monitorit.Register(prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: options.Namespace,
		Subsystem: options.Subsystem,
		Name:      "query_duration_sec",
		Help:      "Histogram of xorm query duration in seconds",
		Buckets:   options.DurationBuckets,
	}, queryLabelNames)).(*prometheus.HistogramVec)

	c.queryCounter = monitorit.Register(prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: options.Namespace,
		Subsystem: options.Subsystem,
		Name:      "query_total",
		Help:      "Number of xorm queries total",
	}, queryLabelNames)).(*prometheus.CounterVec)

	c.errorCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "query_err_total",
			Help: "Total number of xorm query errors",
		},
		errorLabelNames,
	)
	return &c
}

func (h *Hook) BeforeProcess(c *contexts.ContextHook) (context.Context, error) {
	return c.Ctx, nil
}

func (h *Hook) AfterProcess(c *contexts.ContextHook) error {
	queryType := h.getQueryType(c.SQL)
	h.queryCounter.WithLabelValues(h.instanceName, queryType).Inc()
	h.queryHistogram.WithLabelValues(h.instanceName, queryType).Observe(c.ExecuteTime.Seconds())
	if c.Err != nil {
		h.errorCounter.WithLabelValues(h.instanceName, queryType, c.Err.Error()).Inc()
	}
	return nil
}

func (h *Hook) getQueryType(sql string) string {
	index := strings.Index(sql, " ")
	if index != -1 {
		return strings.ToLower(sql[:index])
	} else {
		return "unknown"
	}
}
