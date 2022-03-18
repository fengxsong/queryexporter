package driver

import (
	"github.com/prometheus/client_golang/prometheus"

	"git.irootech.com/sre/queryexporter/pkg/types"
)

type Interface interface {
	Collect(namespace string, m *types.Metric, ch chan<- prometheus.Metric) error
}
