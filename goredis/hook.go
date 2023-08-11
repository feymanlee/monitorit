//
// Package redis
// @Author: feymanlee@gmail.com
// @Description:
// @File:  hook
// @Date: 2023/8/9 15:41
//

package goredis

import (
	"context"
	"fmt"
	"time"

	redis2 "github.com/feymanlee/monitorit/kratos"
	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus"
)

type (
	// Hook represents a go-redis hook that exports metrics of commands and pipelines.
	//
	// The following metrics are exported:
	//
	// - Single commands (not-pipelined)
	//   - Histogram of duration
	//   - Counter of errors
	//
	// - Pipelined commands
	//   - Counter of commands
	//   - Counter of errors
	//
	// The duration of individual pipelined commands won't be collected, but the overall duration of the
	// pipeline will, with a pseudo-command called "pipeline".
	Hook struct {
		options           *Options
		instanceName      string
		singleCommands    *prometheus.HistogramVec
		pipelinedCommands *prometheus.CounterVec
		singleErrors      *prometheus.CounterVec
		pipelinedErrors   *prometheus.CounterVec
	}

	startKey struct{}
)

var (
	commandLabelNames = []string{"instance_name", "command"}
	errorLabelNames   = []string{"instance_name", "command", "error"}
)

// NewHook creates a new go-redis hook instance and registers Prometheus collectors.
func NewHook(instanceName string, opts ...Option) *Hook {
	options := DefaultOptions()
	options.Merge(opts...)
	singleCommands := redis2.Register(prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: options.Namespace,
		Subsystem: options.Subsystem,
		Name:      "single_commands",
		Help:      "Histogram of single Redis commands",
		Buckets:   options.DurationBuckets,
	}, commandLabelNames)).(*prometheus.HistogramVec)

	pipelinedCommands := redis2.Register(prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: options.Namespace,
		Subsystem: options.Subsystem,
		Name:      "pipelined_commands",
		Help:      "Number of pipelined Redis commands",
	}, commandLabelNames)).(*prometheus.CounterVec)

	singleErrors := redis2.Register(prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: options.Namespace,
		Subsystem: options.Subsystem,
		Name:      "single_errors",
		Help:      "Number of single Redis commands that have failed",
	}, errorLabelNames)).(*prometheus.CounterVec)

	pipelinedErrors := redis2.Register(prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: options.Namespace,
		Subsystem: options.Subsystem,
		Name:      "pipelined_errors",
		Help:      "Number of pipelined Redis commands that have failed",
	}, errorLabelNames)).(*prometheus.CounterVec)

	return &Hook{
		options:           options,
		instanceName:      instanceName,
		singleCommands:    singleCommands,
		pipelinedCommands: pipelinedCommands,
		singleErrors:      singleErrors,
		pipelinedErrors:   pipelinedErrors,
	}
}

func (hook *Hook) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	return context.WithValue(ctx, startKey{}, time.Now()), nil
}

func (hook *Hook) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	if start, ok := ctx.Value(startKey{}).(time.Time); ok {
		duration := time.Since(start).Seconds()
		hook.singleCommands.WithLabelValues(hook.instanceName, cmd.Name()).Observe(duration)
	}

	if isActualErr(cmd.Err()) {
		hook.singleErrors.WithLabelValues(hook.instanceName, cmd.Name(), fmt.Sprintf("%v", cmd.Err())).Inc()
	}

	return nil
}

func (hook *Hook) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	return context.WithValue(ctx, startKey{}, time.Now()), nil
}

func (hook *Hook) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
	if err := hook.AfterProcess(ctx, redis.NewCmd(ctx, "pipeline")); err != nil {
		return err
	}

	for _, cmd := range cmds {
		hook.pipelinedCommands.WithLabelValues(hook.instanceName, cmd.Name()).Inc()

		if isActualErr(cmd.Err()) {
			hook.pipelinedErrors.WithLabelValues(hook.instanceName, cmd.Name(), fmt.Sprintf("%v", cmd.Err())).Inc()
		}
	}

	return nil
}

func isActualErr(err error) bool {
	return err != nil && err != redis.Nil
}
