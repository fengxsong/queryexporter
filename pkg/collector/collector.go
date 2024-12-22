package collector

import (
	"context"
	"log/slog"
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/promslog"

	"github.com/fengxsong/queryexporter/pkg/config"
	_ "github.com/fengxsong/queryexporter/pkg/querier"
	"github.com/fengxsong/queryexporter/pkg/querier/factory"
	"github.com/fengxsong/queryexporter/pkg/types"
)

type collector struct {
	namespace string

	cfg                *config.Config
	logger             *slog.Logger
	totalScrapes       *prometheus.CounterVec
	scrapeDurationDesc *prometheus.Desc
	lock               sync.Mutex
}

func New(name string, cfg *config.Config, logger *slog.Logger) (prometheus.Collector, error) {
	if logger == nil {
		logger = promslog.NewNopLogger()
	}
	totalScrapes := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: name,
		Name:      "total_scrapes",
		Help:      "Current total scrapes.",
	}, []string{"driver", "metric", "success"})

	c := &collector{
		namespace:    name,
		cfg:          cfg,
		logger:       logger,
		totalScrapes: totalScrapes,
		scrapeDurationDesc: prometheus.NewDesc(
			prometheus.BuildFQName(name, "", "scrape_duration"),
			"Durations of scrapes",
			[]string{"driver", "metric"}, nil,
		),
	}
	prometheus.MustRegister(c.totalScrapes)

	return c, nil
}

func (c *collector) Describe(_ chan<- *prometheus.Desc) {}

func (c *collector) Collect(ch chan<- prometheus.Metric) {
	c.lock.Lock()
	defer c.lock.Unlock()

	wg := &sync.WaitGroup{}
	ctx := context.Background()

	for driver, metrics := range c.cfg.Aggregations {
		for i := range metrics {
			wg.Add(1)
			go func(subsystem string, a *types.Metric) {
				defer wg.Done()
				start := time.Now()

				err := factory.Default.Process(ctx, c.logger, c.namespace, subsystem, a.DataSources, a.MetricDesc, ch)
				if err != nil {
					c.logger.Error("failed to process", "err", err)
				}
				ch <- prometheus.MustNewConstMetric(
					c.scrapeDurationDesc,
					prometheus.GaugeValue,
					time.Since(start).Seconds(),
					subsystem, a.String())
				c.totalScrapes.WithLabelValues(subsystem, a.String(), strconv.FormatBool(err == nil)).Inc()
			}(string(driver), metrics[i])
		}
	}
	wg.Wait()
}
