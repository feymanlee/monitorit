//
// Package redis
// @Author: feymanlee@gmail.com
// @Description:
// @File:  stat
// @Date: 2023/8/9 16:34
//

package goredis

import (
	"time"

	redis2 "github.com/feymanlee/monitorit/kratos"
	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus"
)

type Stats struct {
	options    *Options
	totalConns prometheus.Gauge
	idleConns  prometheus.Gauge
	staleConns prometheus.Gauge
}

func NewStat(instanceName string, opts ...Option) *Stats {
	options := DefaultOptions()
	options.Merge(opts...)
	stat := Stats{
		options: options,
	}
	statLabels := prometheus.Labels{
		"instance_name": instanceName,
	}
	stat.totalConns = redis2.Register(prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace:   options.Namespace,
		Subsystem:   options.Subsystem,
		Name:        "pool_total_conns",
		Help:        "Number of total connections in the pool",
		ConstLabels: statLabels,
	})).(prometheus.Gauge)

	stat.idleConns = redis2.Register(prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace:   options.Namespace,
		Subsystem:   options.Subsystem,
		Name:        "pool_idle_conns",
		Help:        "Number of idle connections in the pool",
		ConstLabels: statLabels,
	})).(prometheus.Gauge)

	stat.staleConns = redis2.Register(prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace:   options.Namespace,
		Subsystem:   options.Subsystem,
		Name:        "pool_stale_conns",
		Help:        "Number of stale connections removed from the pool",
		ConstLabels: statLabels,
	})).(prometheus.Gauge)

	return &stat
}

func (s *Stats) StartStat(redisClient *redis.Client) {
	go func() {
		for range time.Tick(s.options.StatInterval) {
			stats := redisClient.PoolStats()
			s.totalConns.Set(float64(stats.TotalConns))
			s.idleConns.Set(float64(stats.IdleConns))
			s.staleConns.Set(float64(stats.StaleConns))
		}
	}()
}
