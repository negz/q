package main

import (
	"net"
	"net/http"
	"os"
	"path/filepath"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/negz/q/manager"
	"github.com/negz/q/metrics"
	"github.com/negz/q/rpc"
)

const (
	metricsEndpoint  = "/metrics"
	shutdownEndpoint = "/quitquitquit"
)

func main() {
	var (
		app = kingpin.New(filepath.Base(os.Args[0]), "An in-memory FIFO queue server.")

		debug    = app.Flag("debug", "Run with debug logging.").Short('d').Bool()
		listen   = app.Flag("listen", "Address at which to listen for gRPC connections.").Default(":10002").String()
		listenMx = app.Flag("metrics", "Address at which to expose Prometheus metrics.").Default(":10003").String()
	)
	kingpin.MustParse(app.Parse(os.Args[1:]))

	var log *zap.Logger
	log, err := zap.NewProduction()
	if *debug {
		log, err = zap.NewDevelopment()
	}
	kingpin.FatalIfError(err, "cannot create logger")

	mx, gatherer := metrics.NewPrometheus()
	m := manager.Instrumented(
		manager.New(),
		manager.WithMetrics(mx),
		manager.WithLogger(log),
	)

	l, err := net.Listen("tcp", *listen)
	kingpin.FatalIfError(err, "cannot listen on requested address")
	grpc := rpc.NewServer(l, m)

	r := http.NewServeMux()
	r.Handle(metricsEndpoint, promhttp.HandlerFor(gatherer, promhttp.HandlerOpts{}))
	r.HandleFunc(shutdownEndpoint, func(_ http.ResponseWriter, r *http.Request) {
		log.Info("shutdown requested", zap.String("remote", r.RemoteAddr))
		os.Exit(0)
	})

	e := make(chan error, 1)
	go func(e chan error) {
		e <- grpc.Serve()
	}(e)
	go func(e chan error) {
		e <- http.ListenAndServe(*listenMx, r)
	}(e)
	kingpin.FatalIfError(<-e, "error serving")
}
