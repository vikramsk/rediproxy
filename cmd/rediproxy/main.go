package main

import (
	"errors"
	"flag"
	"log"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/vikramsk/rediproxy/pkg/api"
	"github.com/vikramsk/rediproxy/pkg/cache"
	"github.com/vikramsk/rediproxy/pkg/service"
)

var (
	defaultPort     = "8080"
	defaultRedisURL = "localhost:6379"
	defaultTTL      = time.Hour * 1
	defaultCapacity = 1000000
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		log.Printf("rediproxy: could not start service. err: %+v", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	flagset := flag.NewFlagSet("rediproxy", flag.ExitOnError)
	var (
		port     = flagset.String("port", defaultPort, "proxy service port")
		redisURL = flagset.String("redis-url", defaultRedisURL, "backing redis service address")
		ttl      = flagset.Duration("ttl", defaultTTL, "time to live for cache entries")
		capacity = flagset.Int("capacity", defaultCapacity, "keys limit for the cache")
	)

	if err := flagset.Parse(args); err != nil {
		return err
	}

	_, err := strconv.Atoi(*port)
	if err != nil {
		return errors.New("rediproxy: service port parsing error")
	}

	rc, err := service.NewRedisClient(*redisURL)
	if err != nil {
		return err
	}

	lc := cache.NewLRUCache(*capacity, *ttl)

	pc := service.NewCacheProxy(rc, lc)

	ph := api.NewProxyHandler(pc)

	apiListener, err := net.Listen("tcp", ":"+*port)
	mux := http.NewServeMux()

	// Register pprof handlers
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	mux.Handle("/", ph)

	go interrupt(apiListener)

	log.Printf("launching cache proxy on port: %s", *port)
	log.Fatalf("http server error: %s", http.Serve(apiListener, mux))

	return nil
}

func interrupt(apiListener net.Listener) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	select {
	case sig := <-c:
		log.Printf("received signal: %s", sig)
		apiListener.Close()
	}
}
