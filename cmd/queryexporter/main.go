package main

import (
	"context"
	"fmt"
	"os"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/version"
	"github.com/spf13/pflag"

	"git.irootech.com/sre/queryexporter/pkg/collector"
	"git.irootech.com/sre/queryexporter/pkg/config"
	"git.irootech.com/sre/queryexporter/pkg/handler"
)

const app = "queryexporter"

func main() {
	printVersion := pflag.BoolP("version", "v", false, "Print version info")
	configFile := pflag.StringP("config", "c", "config.yaml", "Configfile path")
	pflag.Parse()
	if *printVersion {
		fmt.Println(version.Print(app))
		os.Exit(0)
	}

	cfg, err := config.ReadFromFile(*configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read config file %s: %v", *configFile, err)
		os.Exit(1)
	}
	logger := initLogger(cfg.LogFormat, cfg.LogLevel)

	clt, err := collector.New(app, cfg, logger)
	if err != nil {
		level.Error(logger).Log("msg", "cannot create collector", "err", err)
		os.Exit(1)
	}
	srv, err := handler.New(
		handler.WithAddr(cfg.Addr),
		handler.WithEnableProfile(*cfg.EnableProfile),
		handler.WithLogger(logger),
		handler.WithCollectors([]prometheus.Collector{clt, version.NewCollector(app)}),
	)
	if err != nil {
		level.Error(logger).Log("err", err)
		os.Exit(1)
	}
	if err = srv.Run(context.Background()); err != nil {
		level.Error(logger).Log("err", err)
		os.Exit(1)
	}
}

func initLogger(format string, lvl string) (logger log.Logger) {
	switch format {
	case "json":
		logger = log.NewJSONLogger(os.Stdout)
	case "console":
		logger = log.NewLogfmtLogger(os.Stdout)
	case "none", "off":
		logger = log.NewNopLogger()
	}
	var lvlOpt level.Option
	switch lvl {
	case "debug":
		lvlOpt = level.AllowAll()
	case "info":
		lvlOpt = level.AllowInfo()
	case "warn", "warning":
		lvlOpt = level.AllowWarn()
	case "error":
		lvlOpt = level.AllowError()
	}
	logger = level.NewFilter(
		log.With(logger, "ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller),
		lvlOpt)
	return logger
}
