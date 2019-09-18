package main

import (
	"net/http"
	"os"

	"github.com/die-net/lrucache/twotier"
	"github.com/gregjones/httpcache"
	"github.com/mikegleasonjr/getcached"
	"github.com/mikegleasonjr/getcached/disk"
	"github.com/mikegleasonjr/getcached/lru"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	version     = "dev"
	listen      = kingpin.Flag("listen", "Listen address (env CP_LISTEN)").Default(":3000").Envar("CP_LISTEN").TCP()
	memsize     = kingpin.Flag("memory-size", "Memory cache size (env CP_MEMORY_SIZE)").Default("25MiB").Envar("CP_MEMORY_SIZE").Bytes()
	diskenabled = kingpin.Flag("enable-disk-cache", "Enable tiered disk cache (env CP_ENABLE_DISK_CACHE)").Default("false").Envar("CP_ENABLE_DISK_CACHE").Default("false").Bool()
	diskdir     = kingpin.Flag("cache-dir", "Cache directory if disk cache enabled (env CP_DISK_CACHE_DIR)").Default(os.TempDir()).PlaceHolder("$TMPDIR").Envar("CP_DISK_CACHE_DIR").ExistingDir()
	disksize    = kingpin.Flag("cache-dir-size", "Disk cache size if disk cache enabled (env CP_DISK_CACHE_SIZE)").Default("100MiB").Envar("CP_DISK_CACHE_SIZE").Bytes()
)

func main() {
	kingpin.Version(version)
	kingpin.Parse()

	memmon, diskmon, cache := configureCaches(uint64(*memsize), *diskenabled, *diskdir, uint64(*disksize))
	prometheus.MustRegister(newCollector("memory", memmon))
	if diskmon != nil {
		prometheus.MustRegister(newCollector("disk", diskmon))
	}

	proxy := getcached.New(getcached.WithCache(cache))
	mux := getMux(proxy)

	log.Infof("getcached:%s started and listening on %s", version, (*listen).String())
	log.Fatal(http.ListenAndServe((*listen).String(), mux))
}

func configureCaches(memsize uint64, diskenabled bool, diskdir string, disksize uint64) (memmon *getcached.Monitor, diskmon *getcached.Monitor, cache httpcache.Cache) {
	memcache := lru.New(lru.WithCache(httpcache.NewMemoryCache()), lru.WithSize(memsize))
	memmon = getcached.NewMonitor(memcache)
	cache = memmon

	if diskenabled {
		diskcache := disk.New(disk.WithDir(diskdir), disk.WithSize(disksize))
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
