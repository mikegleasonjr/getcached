package main

import (
	"github.com/mikegleasonjr/getcached"
	"github.com/prometheus/client_golang/prometheus"
)

type collector struct {
	monitor   *getcached.Monitor
	gets      *prometheus.Desc
	hits      *prometheus.Desc
	hitsBytes *prometheus.Desc
	misses    *prometheus.Desc
	sets      *prometheus.Desc
	setsBytes *prometheus.Desc
	deletes   *prometheus.Desc
}

func newCollector(kind string, monitor *getcached.Monitor) *collector {
	ns := "getcached"
	subs := "cache"
	constLabels := map[string]string{"kind": kind}

	return &collector{
		monitor: monitor,
		gets: prometheus.NewDesc(
			prometheus.BuildFQName(ns, subs, "gets_total"),
			"Total items requested from the cache (misses + hits).",
			nil, constLabels,
		),
		hits: prometheus.NewDesc(
			prometheus.BuildFQName(ns, subs, "gets_hits_total"),
			"Number of cache hits.",
			nil, constLabels,
		),
		hitsBytes: prometheus.NewDesc(
			prometheus.BuildFQName(ns, subs, "gets_hits_bytes"),
			"Total bytes read from the cache.",
			nil, constLabels,
		),
		misses: prometheus.NewDesc(
			prometheus.BuildFQName(ns, subs, "gets_misses_total"),
			"Total cache misses.",
			nil, constLabels,
		),
		sets: prometheus.NewDesc(
			prometheus.BuildFQName(ns, subs, "sets_total"),
			"Total number of items put in the cache.",
			nil, constLabels,
		),
		setsBytes: prometheus.NewDesc(
			prometheus.BuildFQName(ns, subs, "sets_bytes"),
			"Total bytes put in the cache.",
			nil, constLabels,
		),
		deletes: prometheus.NewDesc(
			prometheus.BuildFQName(ns, subs, "deletes_total"),
			"Total number of deletion attemps from the cache.",
			nil, constLabels,
		),
	}
}

func (c *collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.gets
	ch <- c.hits
	ch <- c.hitsBytes
	ch <- c.misses
	ch <- c.sets
	ch <- c.setsBytes
	ch <- c.deletes
}

func (c *collector) Collect(ch chan<- prometheus.Metric) {
	s := c.monitor.Stats()

	// TODO, up metric

	ch <- prometheus.MustNewConstMetric(c.gets, prometheus.CounterValue, float64(s.Gets))
	ch <- prometheus.MustNewConstMetric(c.hits, prometheus.CounterValue, float64(s.Hits))
	ch <- prometheus.MustNewConstMetric(c.hitsBytes, prometheus.CounterValue, float64(s.HitsBytes))
	ch <- prometheus.MustNewConstMetric(c.misses, prometheus.CounterValue, float64(s.Misses))
	ch <- prometheus.MustNewConstMetric(c.sets, prometheus.CounterValue, float64(s.Sets))
	ch <- prometheus.MustNewConstMetric(c.setsBytes, prometheus.CounterValue, float64(s.SetsBytes))
	ch <- prometheus.MustNewConstMetric(c.deletes, prometheus.CounterValue, float64(s.Deletes))
}
