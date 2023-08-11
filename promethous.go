//
// Package monitorit
// @Author: feymanlee@gmail.com
// @Description:
// @File:  promethous
// @Date: 2023/8/11 09:48
//

package monitorit

import "github.com/prometheus/client_golang/prometheus"

func Register(collector prometheus.Collector) prometheus.Collector {
	err := prometheus.DefaultRegisterer.Register(collector)
	if err == nil {
		return collector
	}

	if arErr, ok := err.(prometheus.AlreadyRegisteredError); ok {
		return arErr.ExistingCollector
	}

	panic(err)
}
