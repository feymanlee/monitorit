//
// Package gorm
// @Author: feymanlee@gmail.com
// @Description:
// @File:  options
// @Date: 2023/8/9 20:28
//

package gorm

import "time"

type (
	// Options represents options to customize the exported metrics.
	Options struct {
		Namespace       string
		Subsystem       string
		DurationBuckets []float64
		StatInterval    time.Duration
	}

	Option func(*Options)
)

// DefaultOptions returns the default options.
func DefaultOptions() *Options {
	return &Options{
		Namespace:       "service_component",
		Subsystem:       "gorm",
		DurationBuckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
		StatInterval:    time.Second * 10,
	}
}

func (options *Options) Merge(opts ...Option) {
	for _, opt := range opts {
		opt(options)
	}
}

// WithNamespace sets the namespace of all metrics.
func WithNamespace(namespace string) Option {
	return func(options *Options) {
		options.Namespace = namespace
	}
}

// WithSubsystem sets the namespace of all metrics.
func WithSubsystem(subsystem string) Option {
	return func(options *Options) {
		options.Subsystem = subsystem
	}
}

// WithDurationBuckets sets the duration buckets of single commands metrics.
func WithDurationBuckets(buckets []float64) Option {
	return func(options *Options) {
		options.DurationBuckets = buckets
	}
}

// WithStatInterval sets the duration buckets of single commands metrics.
func WithStatInterval(interval time.Duration) Option {
	return func(options *Options) {
		options.StatInterval = interval
	}
}
