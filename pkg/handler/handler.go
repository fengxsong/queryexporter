package handler

import (
	"context"
	"net"
	"net/http"

	"github.com/go-kit/kit/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Handler struct {
	server *http.Server
}

type options struct {
	addr       string
	logger     log.Logger
	registry   *prometheus.Registry
	collectors []prometheus.Collector
}

type Option func(*options)

func WithAddr(addr string) Option {
	return func(o *options) {
		o.addr = addr
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
	}
	o.registry.MustRegister(o.collectors...)
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.InstrumentMetricHandler(o.registry, promhttp.HandlerFor(o.registry, promhttp.HandlerOpts{})))
	h.server.Handler = mux
	return h, nil
}

func (h *Handler) Run(_ context.Context) error {
	return h.server.ListenAndServe()
}
