package types

import (
	"fmt"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

type Result map[string]interface{}

func (r Result) Get(k string) string {
	val, ok := r[k]
	if !ok {
		return ""
	}
	return fmt.Sprintf("%v", val)
}

func (r Result) GetValue(k string) (float64, error) {
	val, ok := r[k]
	if !ok {
		return 0, fmt.Errorf("cannot find value field %s", k)
	}

	switch val.(type) {
	case float32:
		return float64(val.(float32)), nil
	case float64:
		return val.(float64), nil
	case int32:
		return float64(val.(int32)), nil
	case int64:
		return float64(val.(int64)), nil
	case []byte:
		return strconv.ParseFloat(string(val.([]uint8)), 64)
	default:
		return 0, fmt.Errorf("value must be number /or bytes, type %T value %v given", val, val)
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
