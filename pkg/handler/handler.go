package handler

import (
	"context"
	"net"
	"net/http"
	"net/http/pprof"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Handler struct {
	server *http.Server
	logger log.Logger
}

type options struct {
	addr          string
	enableProfile bool
	logger        log.Logger
	registry      *prometheus.Registry
	collectors    []prometheus.Collector
}

type Option func(*options)

func WithAddr(addr string) Option {
	return func(o *options) {
		o.addr = addr
	}
}

func WithEnableProfile(b bool) Option {
	return func(o *options) {
		o.enableProfile = b
	}
}

func WithLogger(logger log.Logger) Option {
	return func(o *options) {
		o.logger = logger
	}
}

func WithRegistry(registry *prometheus.Registry) Option {
	return func(o *options) {
		o.registry = registry
	}
}

func WithCollectors(collectors []prometheus.Collector) Option {
	return func(o *options) {
		o.collectors = collectors
	}
}

func New(opts ...Option) (*Handler, error) {
	o := &options{
		addr:     ":9696",
		logger:   log.NewNopLogger(),
		registry: prometheus.NewRegistry(),
	}
	for _, f := range opts {
		f(o)
	}
	host, port, err := net.SplitHostPort(o.addr)
	if err != nil {
		return nil, err
	}
	h := &Handler{
		server: &http.Server{
			Addr: net.JoinHostPort(host, port),
		},
		logger: o.logger,
	}
	o.registry.MustRegister(
		prometheus.NewGoCollector(),
		prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}),
	)

	o.registry.MustRegister(o.collectors...)
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.InstrumentMetricHandler(o.registry, promhttp.HandlerFor(o.registry, promhttp.HandlerOpts{})))

	if o.enableProfile {
		level.Info(h.logger).Log("msg", "enable pprofile")
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}

	h.server.Handler = mux
	return h, nil
}

func (h *Handler) Run(_ context.Context) error {
	level.Info(h.logger).Log("msg", "starting HTTP handler", "addr", h.server.Addr)
	return h.server.ListenAndServe()
}
