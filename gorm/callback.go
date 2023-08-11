//
// Package gorm
// @Author: feymanlee@gmail.com
// @Description:
// @File:  callback
// @Date: 2023/8/11 16:27
//

package gorm

import (
	"context"
	"time"

	"github.com/feymanlee/monitorit"
	"github.com/prometheus/client_golang/prometheus"
	"gorm.io/gorm"
)

type contextKey string

var startTimeKey = contextKey("start_time")

type Callback struct {
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

func NewCallback(dbName string, opts ...Option) *Callback {
	options := DefaultOptions()
	options.Merge(opts...)
	c := Callback{
		options:      options,
		instanceName: dbName,
	}
	c.queryHistogram = monitorit.Register(prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: options.Namespace,
		Subsystem: options.Subsystem,
		Name:      "query_duration_sec",
		Help:      "Histogram of GORM query duration in seconds",
		Buckets:   options.DurationBuckets,
	}, queryLabelNames)).(*prometheus.HistogramVec)

	c.queryCounter = monitorit.Register(prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: options.Namespace,
		Subsystem: options.Subsystem,
		Name:      "query_total",
		Help:      "Number of GORM queries total",
	}, queryLabelNames)).(*prometheus.CounterVec)

	c.errorCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "query_err_total",
			Help: "Total number of GORM query errors",
		},
		errorLabelNames,
	)
	return &c
}

func (c *Callback) Register(db *gorm.DB) (err error) {
	// Create
	err = db.Callback().Create().Before("gorm:create").Register("monitor:before_create", c.recordStartTime)
	if err != nil {
		return
	}
	err = db.Callback().Create().After("gorm:create").Register("monitor:after_create", c.recordDurationAndCount("create"))
	if err != nil {
		return
	}

	// Update
	err = db.Callback().Update().Before("gorm:update").Register("monitor:before_update", c.recordStartTime)
	if err != nil {
		return
	}
	err = db.Callback().Update().After("gorm:update").Register("monitor:after_update", c.recordDurationAndCount("update"))
	if err != nil {
		return
	}

	// Delete
	err = db.Callback().Delete().Before("gorm:delete").Register("monitor:before_delete", c.recordStartTime)
	if err != nil {
		return
	}
	err = db.Callback().Delete().After("gorm:delete").Register("monitor:after_delete", c.recordDurationAndCount("delete"))
	if err != nil {
		return
	}

	// Query
	err = db.Callback().Query().Before("gorm:query").Register("monitor:before_query", c.recordStartTime)
	if err != nil {
		return
	}
	err = db.Callback().Query().After("gorm:query").Register("monitor:after_query", c.recordDurationAndCount("query"))
	return
}

func (c *Callback) recordStartTime(db *gorm.DB) {
	ctx := db.Statement.Context
	newCtx := context.WithValue(ctx, startTimeKey, time.Now())
	db.Statement.Context = newCtx
}

func (c *Callback) recordDurationAndCount(queryType string) func(*gorm.DB) {
	return func(db *gorm.DB) {
		startTimeValue := db.Statement.Context.Value(startTimeKey)
		startTime, ok := startTimeValue.(time.Time)
		if !ok {
			return
		}
		duration := time.Since(startTime).Seconds()
		c.queryCounter.WithLabelValues(c.instanceName, queryType).Inc()
		c.queryHistogram.WithLabelValues(c.instanceName, queryType).Observe(duration)

		// If there was an error, increment the error counter with the error reason
		if db.Error != nil {
			c.errorCounter.WithLabelValues(c.instanceName, queryType, db.Error.Error()).Inc()
		}
	}
}
