package types

import (
	"fmt"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cast"
)

type Result map[string]interface{}

func (r Result) Get(k string) string {
	val, ok := r[k]
	if !ok {
		return ""
	}
	return cast.ToString(val)
}

func (r Result) GetValue(k string) (float64, error) {
	val, ok := r[k]
	if !ok {
		return 0, fmt.Errorf("cannot find value field %s", k)
	}
	switch k := val.(type) {
	case []byte:
		return strconv.ParseFloat(string(k), 64)
	default:
		return cast.ToFloat64E(val)
	}
}

var builtinLabels = []string{"name", "database", "table"}

func CreateMetric(namespace, subsystem string, ds *DataSource, metric *Metric, ret Result) (prometheus.Metric, error) {
	val, err := ret.GetValue(metric.VariableValue)
	if err != nil {
		return nil, err
	}
	labelValues := make([]string, 0)
	desc := metric.Desc(namespace, subsystem, builtinLabels...)
	for _, labelVar := range metric.VariableLabels {
		labelValues = append(labelValues, ret.Get(labelVar))
	}
	labelValues = append(labelValues, ds.Name, ds.Database, ds.Table)
	return prometheus.NewConstMetric(desc, prometheus.GaugeValue, val, labelValues...)
}
