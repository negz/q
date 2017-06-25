package main

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/facebookgo/httpdown"
	pb "github.com/golang/protobuf/proto"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/negz/q/rpc/proto"
)

const (
	metricsEndpoint  = "/metrics"
	shutdownEndpoint = "/quitquitquit"
)

type loggingHandler struct {
	h   http.Handler
	log *zap.Logger
}

func (h *loggingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.log.Info("request",
		zap.String("host", r.Host),
		zap.String("method", r.Method),
		zap.String("url", r.URL.String()),
		zap.String("agent", r.UserAgent()),
		zap.String("remote", r.RemoteAddr))
	h.h.ServeHTTP(w, r)
}

func logForwardedResponse(log *zap.Logger) func(context.Context, http.ResponseWriter, pb.Message) error {
	return func(_ context.Context, _ http.ResponseWriter, m pb.Message) error {
		log.Debug("forwarded response", zap.Any("message", m))
		return nil
	}
}

func main() {
	var (
		app = kingpin.New(filepath.Base(os.Args[0]), "REST API gateway for in-memory FIFO queue server.").DefaultEnvars()

		server = app.Arg("server", "Address at which to query queue server.").String()

		listen = app.Flag("listen", "Address at which to listen.").Default(":80").String()
		debug  = app.Flag("debug", "Run with debug logging.").Short('d').Bool()
		stop   = app.Flag("close-after", "Wait this long at shutdown before closing HTTP connections.").Default("1m").Duration()
		kill   = app.Flag("kill-after", "Wait this long at shutdown before exiting.").Default("2m").Duration()
	)

	kingpin.MustParse(app.Parse(os.Args[1:]))

	var log *zap.Logger
	log, err := zap.NewProduction()
	if *debug {
		log, err = zap.NewDevelopment()
	}
	kingpin.FatalIfError(err, "cannot create log")
	log.Info("gateway initialised", zap.String("listen", *listen), zap.String("server", *server))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	gw := runtime.NewServeMux(runtime.WithForwardResponseOption(logForwardedResponse(log)))
	kingpin.FatalIfError(proto.RegisterQHandlerFromEndpoint(ctx, gw, *server, []grpc.DialOption{grpc.WithInsecure()}), "cannot register REST handler")

	r := http.NewServeMux()
	r.Handle("/", gw)
	r.Handle(metricsEndpoint, promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{}))
	r.HandleFunc(shutdownEndpoint, func(_ http.ResponseWriter, r *http.Request) {
		log.Info("shutdown requested", zap.String("remote", r.RemoteAddr))
		os.Exit(0)
	})

	hd := &httpdown.HTTP{StopTimeout: *stop, KillTimeout: *kill}
	http := &http.Server{Addr: *listen, Handler: &loggingHandler{cors.Default().Handler(r), log}}
	kingpin.FatalIfError(httpdown.ListenAndServe(http, hd), "HTTP server error")
}
