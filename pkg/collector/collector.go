package collector

import (
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/prometheus/client_golang/prometheus"

	"git.irootech.com/sre/queryexporter/pkg/config"
	"git.irootech.com/sre/queryexporter/pkg/types"
)

type queries struct {
	namespace string
	cfg       config.Config
	logger    log.Logger

	scrapeDurationDesc *prometheus.Desc
	lock               sync.Mutex
}

func New(name string, cfg *config.Config, logger log.Logger) (prometheus.Collector, error) {
	return nil, nil
}

func (q *queries) Describe(ch chan<- *prometheus.Desc) {
	ch <- q.scrapeDurationDesc
}

func (q *queries) Collect(ch chan<- prometheus.Metric) {
	q.lock.Lock()
	defer q.lock.Unlock()

	wg := &sync.WaitGroup{}

	for driver, metrics := range q.cfg.Metrics {
		for i := range metrics {
			wg.Add(1)
			go func(namespace string, metric *types.Metric) {
				defer wg.Done()
				start := time.Now()
				// TODO: do actual collect
				ch <- prometheus.MustNewConstMetric(
					q.scrapeDurationDesc,
					prometheus.GaugeValue,
					time.Now().Sub(start).Seconds(),
					namespace, metric.String())
			}(driver, metrics[i])
		}
	}
	wg.Wait()
}
