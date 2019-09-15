package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/die-net/lrucache/twotier"
	"github.com/gregjones/httpcache"
	"github.com/mikegleasonjr/getcached"
	"github.com/mikegleasonjr/getcached/disk"
	"github.com/mikegleasonjr/getcached/lru"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	version     = "dev"
	listen      = kingpin.Flag("listen", "Listen address (env CP_LISTEN)").Default(":3000").Envar("CP_LISTEN").TCP()
	memsize     = kingpin.Flag("memory-size", "Memory cache size (env CP_MEMORY_SIZE)").Default("25MiB").Envar("CP_MEMORY_SIZE").Bytes()
	diskenabled = kingpin.Flag("enable-cache-dir", "Enable tiered disk cache (env CP_ENABLE_DISK_CACHE)").Default("false").Envar("CP_ENABLE_DISK_CACHE").Default("false").Bool()
	diskdir     = kingpin.Flag("cache-dir", "Cache directory if disk cache enabled (env CP_DISK_CACHE_DIR)").Default(os.TempDir()).PlaceHolder("$TMPDIR").Envar("CP_DISK_CACHE_DIR").ExistingDir()
	disksize    = kingpin.Flag("cache-dir-size", "Disk cache size if disk cache enabled (env CP_DISK_CACHE_SIZE)").Default("100MiB").Envar("CP_DISK_CACHE_SIZE").Bytes()
)

func main() {
	kingpin.Version(version)
	kingpin.Parse()

	var memmon *getcached.Monitor
	var diskmon *getcached.Monitor
	var cache httpcache.Cache

	memcache := lru.New(lru.WithCache(httpcache.NewMemoryCache()), lru.WithSize(uint64(*memsize)))
	memmon = getcached.NewMonitor(memcache)
	cache = memmon
	if *diskenabled {
		diskcache := disk.New(disk.WithDir(*diskdir), disk.WithSize(uint64(*disksize)))
		diskmon = getcached.NewMonitor(diskcache)
		cache = twotier.New(memmon, diskmon)
	}

	go func() {
		t := time.NewTicker(time.Second * 5)
		for {
			<-t.C
			fmt.Printf("mem : %#v\n", memmon.Stats())
			if diskmon != nil {
				fmt.Printf("disk: %#v\n", diskmon.Stats())
			}
		}
	}()

	proxy := getcached.New(getcached.WithCache(cache))
	log.Printf("started getcached:%s on %s", version, (*listen).String())
	log.Fatal(http.ListenAndServe((*listen).String(), proxy))
}
