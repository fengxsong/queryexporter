package factory

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
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

var bufPool = sync.Pool{
	New: func() any {
		return &bytes.Buffer{}
	},
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
			l := log.With(logger, "driver", driver)
			rets, err := iface.Query(logutil.InjectContext(ctx, l), ds, buf.String())
			if err != nil {
				return err
			}
			level.Debug(l).Log("results", rets)
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

var defaultTpl *template.Template

func init() {
	defaultTpl = template.New("goTpl").
		Option("missingkey=default").
		Funcs(sprig.TxtFuncMap())
}
