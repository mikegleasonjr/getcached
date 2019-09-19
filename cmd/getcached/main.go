package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/die-net/lrucache/twotier"
	"github.com/gregjones/httpcache"
	"github.com/mikegleasonjr/getcached"
	"github.com/mikegleasonjr/getcached/disk"
	"github.com/mikegleasonjr/getcached/lru"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	version     = "v0.0.0-dev"
	stderr      = log.New(os.Stderr, "[getcached] ", log.LstdFlags|log.Lshortfile)
	stdout      = log.New(os.Stdout, "[getcached] ", log.LstdFlags)
	listen      = kingpin.Flag("listen", "Listen address (env CP_LISTEN)").Default(":3000").Envar("CP_LISTEN").TCP()
	memsize     = kingpin.Flag("memory-size", "Memory cache size (env CP_MEMORY_SIZE)").Default("25MiB").Envar("CP_MEMORY_SIZE").Bytes()
	diskenabled = kingpin.Flag("enable-disk-cache", "Enable tiered disk cache (env CP_ENABLE_DISK_CACHE)").Default("false").Envar("CP_ENABLE_DISK_CACHE").Default("false").Bool()
	diskdir     = kingpin.Flag("cache-dir", "Cache directory if disk cache enabled (env CP_DISK_CACHE_DIR)").Default(os.TempDir()).PlaceHolder("$TMPDIR").Envar("CP_DISK_CACHE_DIR").ExistingDir()
	disksize    = kingpin.Flag("cache-dir-size", "Disk cache size if disk cache enabled (env CP_DISK_CACHE_SIZE)").Default("100MiB").Envar("CP_DISK_CACHE_SIZE").Bytes()
	maxbodysize = kingpin.Flag("max-body-size", "Max response body size allowed to be downloaded (env CP_MAX_BODY_SIZE)").Default("10MiB").Envar("CP_MAX_BODY_SIZE").Bytes()
)

func main() {
	kingpin.Version(version)
	kingpin.Parse()

	memmon, diskmon, cache := configureCaches(uint64(*memsize), *diskenabled, *diskdir, uint64(*disksize))
	proxy := getcached.New(
		getcached.WithCache(cache),
		getcached.WithBufferPool(getcached.DefaultBufferPool),
		getcached.WithErrorLogger(stderr),
		getcached.WithProxyTransport(BodySizeCheckerTransport(int64(*maxbodysize), http.DefaultTransport)),
	)
	mux := getMux(proxy)
	registerPrometheusMetrics(memmon, diskmon)

	stdout.Printf("%s listening on %s", version, (*listen).String())
	stderr.Println(gracefulServe((*listen).String(), mux))
}

func configureCaches(memsize uint64, diskenabled bool, diskdir string, disksize uint64) (memmon *getcached.Monitor, diskmon *getcached.Monitor, cache httpcache.Cache) {
	memcache := lru.New(lru.WithCache(httpcache.NewMemoryCache()), lru.WithSize(memsize))
	memmon = getcached.NewMonitor(memcache)
	cache = memmon

	if diskenabled {
		diskcache := lru.New(lru.WithCache(disk.New(disk.WithDir(diskdir))), lru.WithSize(disksize))
		diskmon = getcached.NewMonitor(diskcache)
		cache = twotier.New(memmon, diskmon)
	}

	return
}

func getMux(proxy *getcached.Proxy) *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("/metrics", promhttp.Handler())
	mux.Handle("/", proxy)

	return mux
}

func registerPrometheusMetrics(mem, disk *getcached.Monitor) {
	metrics := newMetrics()
	metrics.addCollector("memory", mem)
	if disk != nil {
		metrics.addCollector("disk", disk)
	}
	prometheus.MustRegister(metrics)
}

func gracefulServe(addr string, handler http.Handler) error {
	sigchan := make(chan os.Signal)
	res := make(chan error, 1)
	srv := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		res <- srv.ListenAndServe()
	}()

	select {
	case <-sigchan:
	case err := <-res:
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), srv.WriteTimeout)
	defer cancel()
	srv.Shutdown(ctx)
	return <-res
}
