package collector

import (
	"context"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"

	"git.irootech.com/sre/queryexporter/pkg/config"
	_ "git.irootech.com/sre/queryexporter/pkg/querier"
	"git.irootech.com/sre/queryexporter/pkg/querier/factory"
)

type queries struct {
	namespace string
	cfg       *config.Config
	logger    log.Logger

	scrapeDurationDesc *prometheus.Desc
	lock               sync.Mutex
}

func New(name string, cfg *config.Config, logger log.Logger) (prometheus.Collector, error) {
	if logger == nil {
		logger = log.NewNopLogger()
	}
	qs := &queries{
		namespace: name,
		cfg:       cfg,
		logger:    logger,
		scrapeDurationDesc: prometheus.NewDesc(
			prometheus.BuildFQName(name, "", "scrape_duration"),
			"querier scrape duration",
			[]string{"driver", "metric"}, nil,
		),
	}
	return qs, nil
}

func (q *queries) Describe(ch chan<- *prometheus.Desc) {
	ch <- q.scrapeDurationDesc
}

func (q *queries) Collect(ch chan<- prometheus.Metric) {
	q.lock.Lock()
	defer q.lock.Unlock()

	wg := &sync.WaitGroup{}
	ctx := context.Background()

	for driver, metrics := range q.cfg.Metrics {
		for i := range metrics {
			wg.Add(1)
			go func(subsystem string, metric *config.Metric) {
				defer wg.Done()
				start := time.Now()
				// TODO: do actual collect
				err := factory.Default.Process(ctx, q.namespace, driver, metric.DataSources, metric.Metric, ch)
				if err != nil {
					level.Error(q.logger).Log("err", err)
				}
				ch <- prometheus.MustNewConstMetric(
					q.scrapeDurationDesc,
					prometheus.GaugeValue,
					time.Now().Sub(start).Seconds(),
					subsystem, metric.String())
			}(driver, metrics[i])
		}
	}
	wg.Wait()
}
