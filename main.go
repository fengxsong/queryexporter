package main

import (
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/client_golang/prometheus"
	versioncollector "github.com/prometheus/client_golang/prometheus/collectors/version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promslog"
	"github.com/prometheus/common/promslog/flag"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
	webflag "github.com/prometheus/exporter-toolkit/web/kingpinflag"

	"github.com/fengxsong/queryexporter/pkg/collector"
	"github.com/fengxsong/queryexporter/pkg/config"
)

const app = "queryexporter"

func init() {
	prometheus.MustRegister(versioncollector.NewCollector(app))
}

func main() {
	os.Exit(run())
}

func newLandingPage(metricsPath, healthzPath string) (http.Handler, error) {
	landingConfig := web.LandingConfig{
		Name:        app,
		Description: "exporter for many database sources",
		Version:     version.Info(),
		Links: []web.LandingLinks{
			{
				Address:     metricsPath,
				Text:        "Metrics",
				Description: "for self-metrics or running in single target mode",
			},
			{
				Address:     healthzPath,
				Text:        "Healthz",
				Description: "for liveness or readiness probe",
			},
		},
	}

	return web.NewLandingPage(landingConfig)
}

func run() int {
	var (
		toolkitFlags = webflag.AddFlags(kingpin.CommandLine, ":9696")

		metricsPath = kingpin.Flag(
			"web.telemetry-path",
			"Path under which to expose metrics.").Default("/metrics").String()
		configF   = kingpin.Flag("config", "Path of config file").Short('c').Default("config.yaml").String()
		expandEnv = kingpin.Flag("expand-env", "Expand env in config file, for reading secrets from environment variables").Default("false").Bool()
		test      = kingpin.Flag("test", "Print rendered content of config file").Short('t').Default("false").Bool()
		namespace = kingpin.Flag("namespace", "Namespace for metrics").Short('n').Default(app).String()
	)
	promslogConfig := &promslog.Config{}

	flag.AddFlags(kingpin.CommandLine, promslogConfig)
	kingpin.Version(version.Print(app))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	logger := promslog.New(promslogConfig)

	cfg, err := config.ReadFromFile(*configF, *expandEnv)
	if err != nil {
		logger.Error("failed to read config", "err", err)
		return 1
	}

	if *test {
		config.Dump(cfg, os.Stdout)
		return 0
	}
	c, err := collector.New(*namespace, cfg, logger)
	if err != nil {
		logger.Error("failed to create collector", "err", err)
		return 1
	}

	prometheus.MustRegister(c)

	http.Handle(*metricsPath, promhttp.Handler())

	healthzPath := "/-/healthy"
	http.HandleFunc(healthzPath, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Healthy"))
	})

	landingPage, err := newLandingPage(*metricsPath, healthzPath)
	if err != nil {
		logger.Error("failed to create landing page", "err", err)
		return 1
	}

	http.Handle("/", landingPage)

	srv := &http.Server{}
	srvc := make(chan struct{})
	term := make(chan os.Signal, 1)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := web.ListenAndServe(srv, toolkitFlags, logger); err != nil {
			logger.Error("error starting HTTP server", "err", err)
			close(srvc)
		}
	}()

	for {
		select {
		case <-term:
			logger.Info("received SIGTERM, exiting gracefully")
			return 0
		case <-srvc:
			return 1
		}
	}
}
