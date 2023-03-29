package factory

import (
	"context"
	"fmt"

	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"

	logutil "github.com/fengxsong/queryexporter/pkg/logger"
	"github.com/fengxsong/queryexporter/pkg/types"
)

type Interface interface {
	// Name() string
	Query(ctx context.Context, ds *types.DataSource, query string) ([]types.Result, error)
}

type Factory struct {
	queriers map[string]Interface
}

func (f *Factory) Process(ctx context.Context, logger log.Logger, namespace, driver string, dss []*types.DataSource, metric *types.Metric, ch chan<- prometheus.Metric) error {
	eg, ctx := errgroup.WithContext(ctx)
	for i := range dss {
		ds := dss[i]
		eg.Go(func() error {
			iface, ok := f.queriers[driver]
			if !ok {
				return fmt.Errorf("querier %s not implemented yet", driver)
			}
			l := log.With(logger, "driver", driver)
			rets, err := iface.Query(logutil.InjectContext(ctx, l), ds, metric.Query)
			if err != nil {
				return err
			}
			for i := range rets {
				m, err := types.CreateMetric(namespace, driver, ds, metric, rets[i])
				if err != nil {
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
