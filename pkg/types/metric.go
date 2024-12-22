package types

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	TypeGauge = "gauge"
)

type MetricDesc struct {
	Name           string            `json:"name"`
	Help           string            `json:"help"`
	Type           string            `json:"type" default:"gauge"`
	VariableValue  string            `json:"variableValue"` // for taking value from result
	Query          string            `json:"query"`
	VariableLabels []string          `json:"variableLabels,omitempty"` // for dynamic labels from query results
	ConstLabels    prometheus.Labels `json:"constLabels,omitempty"`

	desc *prometheus.Desc
}

func (m *MetricDesc) String() string {
	return m.Name
}

func (m *MetricDesc) setDefaults() {}

func (m *MetricDesc) Validate() error {
	switch m.Type {
	case TypeGauge, "":
	default:
		return fmt.Errorf("unsupported type %s", m.Type)
	}
	if m.VariableValue == "" {
		return fmt.Errorf("variableValue field must specified for metric %s", m.Name)
	}
	return nil
}

func (m *MetricDesc) ToDesc(namespace, subsystem string, labels ...string) *prometheus.Desc {
	if len(labels) < 3 {
		panic("Must include builtin labels name/database/table")
	}
	variableLabels := append(m.VariableLabels, labels...)
	if m.desc == nil {
		m.desc = prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, m.Name),
			m.Help, variableLabels, m.ConstLabels,
		)
	}
	return m.desc
}
