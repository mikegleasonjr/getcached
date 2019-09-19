package main

import (
	"github.com/mikegleasonjr/getcached"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	ns       = "getcached"
	subs     = "cache"
	locLabel = "loc"
)

var m *metrics

type metrics struct {
	collectors []*collector
	up         *prometheus.Desc
}

func newMetrics() *metrics {
	return &metrics{
		up: prometheus.NewDesc(
			prometheus.BuildFQName(ns, "", "up"),
			"Is getcached up?",
			nil, nil,
		),
	}
}

func (m *metrics) addCollector(loc string, monitor *getcached.Monitor) {
	m.collectors = append(m.collectors, newCollector(loc, monitor))
}

func (m *metrics) Describe(ch chan<- *prometheus.Desc) {
	ch <- m.up
	for _, c := range m.collectors {
		c.Describe(ch)
	}
}

func (m *metrics) Collect(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(m.up, prometheus.GaugeValue, 1)
	for _, c := range m.collectors {
		c.Collect(ch)
	}
}

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

func newCollector(loc string, monitor *getcached.Monitor) *collector {
	constLabels := map[string]string{locLabel: loc}

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

	ch <- prometheus.MustNewConstMetric(c.gets, prometheus.CounterValue, float64(s.Gets))
	ch <- prometheus.MustNewConstMetric(c.hits, prometheus.CounterValue, float64(s.Hits))
	ch <- prometheus.MustNewConstMetric(c.hitsBytes, prometheus.CounterValue, float64(s.HitsBytes))
	ch <- prometheus.MustNewConstMetric(c.misses, prometheus.CounterValue, float64(s.Misses))
	ch <- prometheus.MustNewConstMetric(c.sets, prometheus.CounterValue, float64(s.Sets))
	ch <- prometheus.MustNewConstMetric(c.setsBytes, prometheus.CounterValue, float64(s.SetsBytes))
	ch <- prometheus.MustNewConstMetric(c.deletes, prometheus.CounterValue, float64(s.Deletes))
}
