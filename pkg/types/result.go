package types

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cast"
)

type Result map[string]any

func (r Result) Get(k string) string {
	val := jsonPathGet(r, k)
	if val == nil {
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

func (r Result) IsEmpty() bool {
	return len(r) == 0
}

func jsonPathGet(objects map[string]any, key string) any {
	// fastest path
	if v, ok := objects[key]; ok {
		return v
	}
	idx := strings.Index(key, ".")
	if idx < 0 {
		return objects[key]
	}

	if v, ok := objects[key[0:idx]]; ok {
		if m, ok1 := v.(map[string]any); ok1 {
			return jsonPathGet(m, key[idx+1:])
		}
		// if v is an object that may implements json.Marshaler, we need to parse it
		// then revert it to map[string]any
		parsed, err := json.Marshal(v)
		if err == nil {
			var obj map[string]any
			err := json.Unmarshal(parsed, &obj)
			if err == nil {
				return jsonPathGet(obj, key[idx+1:])
			}
		}
		return v
	}
	return nil
}

var builtinLabels = []string{"name", "database", "table"}

func CreateGaugeMetric(namespace, subsystem string, ds *DataSource, m *MetricDesc, ret Result) (prometheus.Metric, error) {
	var (
		val float64
		err error
	)
	if ret.IsEmpty() && m.AllowEmptyValue {
		val = 0
	} else {
		val, err = ret.GetValue(m.VariableValue)
	}
	if err != nil {
		return nil, err
	}
	labelValues := make([]string, 0, len(m.VariableLabels)+len(builtinLabels))
	desc := m.ToDesc(namespace, subsystem, builtinLabels...)
	for _, labelVar := range m.VariableLabels {
		labelValues = append(labelValues, ret.Get(labelVar))
	}
	labelValues = append(labelValues, ds.Name, ds.Database, ds.Table)
	return prometheus.NewConstMetric(desc, prometheus.GaugeValue, val, labelValues...)
}
