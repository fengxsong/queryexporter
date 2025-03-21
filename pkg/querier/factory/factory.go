package factory

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"sync"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"

	"github.com/fengxsong/queryexporter/pkg/querier/log"
	"github.com/fengxsong/queryexporter/pkg/types"
)

type Interface interface {
	Query(ctx context.Context, ds *types.DataSource, query string) ([]types.Result, error)
}

type Factory struct {
	queriers map[string]Interface
}

var bufPool = sync.Pool{
	New: func() any {
		return &bytes.Buffer{}
	},
}

func (f *Factory) Process(ctx context.Context, logger *slog.Logger, namespace, driver string, dss []*types.DataSource, metric *types.MetricDesc, ch chan<- prometheus.Metric) error {
	logger = logger.With("driver", driver)
	eg, ctx := errgroup.WithContext(ctx)
	for i := range dss {
		ds := dss[i]
		eg.Go(func() error {
			iface, ok := f.queriers[driver]
			if !ok {
				return fmt.Errorf("querier %s not implemented yet", driver)
			}
			tp, err := defaultTpl.Clone()
			if err != nil {
				return err
			}
			tp, err = tp.Parse(metric.Query)
			if err != nil {
				return err
			}
			buf := bufPool.Get().(*bytes.Buffer)
			defer func() {
				buf.Reset()
				bufPool.Put(buf)
			}()
			// currently only support using some template functions
			if err = tp.Execute(buf, map[string]any{}); err != nil {
				return err
			}

			ctx = log.WithLogger(ctx, logger)

			rets, err := iface.Query(ctx, ds, buf.String())
			if err != nil {
				if metric.ContinueIfError {
					logger.Error("failed to query", "datasource", dss, "metric", metric.String(), "err", err)
					return nil
				}
				return fmt.Errorf("failed to query %s with %s, err: %v", ds.String(), buf.String(), err)
			}
			logger.With("driver", driver).Debug("",
				"datasource", dss, "metric", metric.String(),
				"results", rets)
			for i := range rets {
				m, err := types.CreateGaugeMetric(namespace, driver, ds, metric, rets[i])
				if err != nil {
					if metric.ContinueIfError {
						logger.Error("failed to create metric", "datasource", dss, "metric", metric.String(), "err", err)
						continue
					}
					return err
				}
				ch <- m
			}
			return nil
		})
	}
	return eg.Wait()
}

func (f *Factory) Register(driver string, iface Interface) {
	if _, ok := Default.queriers[driver]; ok {
		panic(fmt.Sprintf("driver %s duplicated", driver))
	}
	Default.queriers[driver] = iface
}

var Default = &Factory{
	queriers: make(map[string]Interface),
}

func Register(driver string, iface Interface) {
	Default.Register(driver, iface)
}

var defaultTpl *template.Template

func init() {
	defaultTpl = template.New("goTpl").
		Option("missingkey=default").
		Funcs(sprig.TxtFuncMap())
}
