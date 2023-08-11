//
// Package gorm
// @Author: feymanlee@gmail.com
// @Description:
// @File:  stat
// @Date: 2023/8/9 20:28
//

package gorm

import (
	"log"
	"time"

	"github.com/feymanlee/monitorit"
	"github.com/prometheus/client_golang/prometheus"
	"gorm.io/gorm"
)

type DBStats struct {
	options            *Options
	maxOpenConnections prometheus.Gauge // Maximum number of open connections to the database.

	// Pool status
	openConnections prometheus.Gauge // The number of established connections both in use and idle.
	inUse           prometheus.Gauge // The number of connections currently in use.
	idle            prometheus.Gauge // The number of idle connections.

	// Counters
	waitCount         prometheus.Gauge // The total number of connections waited for.
	waitDuration      prometheus.Gauge // The total time blocked waiting for a new connection.
	maxIdleClosed     prometheus.Gauge // The total number of connections closed due to SetMaxIdleConns.
	maxLifetimeClosed prometheus.Gauge // The total number of connections closed due to SetConnMaxLifetime.
	maxIdleTimeClosed prometheus.Gauge // The total number of connections closed due to SetConnMaxIdleTime.
}

func NewStats(dbName string, opts ...Option) *DBStats {
	statLabelNames := prometheus.Labels{
		"db_name": dbName,
	}
	options := DefaultOptions()
	options.Merge(opts...)
	stats := &DBStats{
		options: options,
		maxOpenConnections: monitorit.Register(prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: options.Namespace,
			Subsystem: options.Subsystem,
			Name:      "dbstats_max_open_connections",

			Help:        "Maximum number of open connections to the database.",
			ConstLabels: statLabelNames,
		})).(prometheus.Gauge),
		openConnections: monitorit.Register(prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   options.Namespace,
			Subsystem:   options.Subsystem,
			Name:        "dbstats_open_connections",
			Help:        "The number of established connections both in use and idle.",
			ConstLabels: statLabelNames,
		})).(prometheus.Gauge),
		inUse: monitorit.Register(prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   options.Namespace,
			Subsystem:   options.Subsystem,
			Name:        "dbstats_in_use",
			Help:        "The number of connections currently in use.",
			ConstLabels: statLabelNames,
		})).(prometheus.Gauge),
		idle: monitorit.Register(prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   options.Namespace,
			Subsystem:   options.Subsystem,
			Name:        "dbstats_idle",
			Help:        "The number of idle connections.",
			ConstLabels: statLabelNames,
		})).(prometheus.Gauge),
		waitCount: monitorit.Register(prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   options.Namespace,
			Subsystem:   options.Subsystem,
			Name:        "dbstats_wait_count",
			Help:        "The total number of connections waited for.",
			ConstLabels: statLabelNames,
		})).(prometheus.Gauge),
		waitDuration: monitorit.Register(prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   options.Namespace,
			Subsystem:   options.Subsystem,
			Name:        "dbstats_wait_duration",
			Help:        "The total time blocked waiting for a new connection.",
			ConstLabels: statLabelNames,
		})).(prometheus.Gauge),
		maxIdleClosed: monitorit.Register(prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   options.Namespace,
			Subsystem:   options.Subsystem,
			Name:        "dbstats_max_idle_closed",
			Help:        "The total number of connections closed due to SetMaxIdleConns.",
			ConstLabels: statLabelNames,
		})).(prometheus.Gauge),
		maxLifetimeClosed: monitorit.Register(prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   options.Namespace,
			Subsystem:   options.Subsystem,
			Name:        "dbstats_max_lifetime_closed",
			Help:        "The total number of connections closed due to SetConnMaxLifetime.",
			ConstLabels: statLabelNames,
		})).(prometheus.Gauge),
		maxIdleTimeClosed: monitorit.Register(prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   options.Namespace,
			Subsystem:   options.Subsystem,
			Name:        "dbstats_max_idletime_closed",
			Help:        "The total number of connections closed due to SetConnMaxIdleTime.",
			ConstLabels: statLabelNames,
		})).(prometheus.Gauge),
	}

	return stats
}

func (s *DBStats) StartStats(db *gorm.DB) {
	go func() {
		for range time.Tick(s.options.StatInterval) {
			if dba, err := db.DB(); err == nil {
				dbStats := dba.Stats()
				s.maxOpenConnections.Set(float64(dbStats.MaxOpenConnections))
				s.openConnections.Set(float64(dbStats.OpenConnections))
				s.inUse.Set(float64(dbStats.InUse))
				s.idle.Set(float64(dbStats.Idle))
				s.waitCount.Set(float64(dbStats.WaitCount))
				s.waitDuration.Set(float64(dbStats.WaitDuration))
				s.maxIdleClosed.Set(float64(dbStats.MaxIdleClosed))
				s.maxLifetimeClosed.Set(float64(dbStats.MaxLifetimeClosed))
				s.maxIdleTimeClosed.Set(float64(dbStats.MaxIdleTimeClosed))
			} else {
				log.Printf("gorm:prometheus failed to collect db status, got error: %v", err)
			}
		}
	}()
}
