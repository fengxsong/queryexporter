package factory

import (
	"context"
	"fmt"
	"sync"

	"github.com/prometheus/client_golang/prometheus"

	"git.irootech.com/sre/queryexporter/pkg/types"
)

type Interface interface {
	// Name() string
	Query(ctx context.Context, ds *types.DataSource, metric *types.Metric) ([]types.Result, error)
}

type Factory struct {
	queriers map[string]Interface
}

func (f *Factory) Process(ctx context.Context, namespace, driver string, dss []*types.DataSource, metric *types.Metric, ch chan<- prometheus.Metric) error {
	wg := &sync.WaitGroup{}
	errCh := make(chan error, 1)
	for i := range dss {
		wg.Add(1)
		go func(ds *types.DataSource) {
			defer wg.Done()
			iface, ok := f.queriers[driver]
			if !ok {
				errCh <- fmt.Errorf("querier %s not implemented yet", driver)
				return
			}
			rets, err := iface.Query(ctx, ds, metric)
			if err != nil {
				errCh <- err
				return
			}
			for i := range rets {
				m, err := types.CreateMetric(namespace, driver, ds, metric, rets[i])
				if err != nil {
					errCh <- err
					return
				}
				ch <- m
			}
		}(dss[i])
	}
	wg.Wait()
	select {
	case err := <-errCh:
		return err
	default:
		return nil
	}
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
