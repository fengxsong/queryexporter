package types

import (
	"fmt"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	TypeGauge = "gauge"
)

type MetricDesc struct {
	Name            string            `json:"name"`
	Help            string            `json:"help"`
	Type            string            `json:"type" default:"gauge"`
	VariableValue   string            `json:"variableValue"` // for taking value from result
	Query           string            `json:"query"`
	VariableLabels  []string          `json:"variableLabels,omitempty"` // for dynamic labels from query results
	ConstLabels     prometheus.Labels `json:"constLabels,omitempty"`
	ContinueIfError bool              `json:"continueIfError,omitempty"`
	AllowEmptyValue bool              `json:"allowEmptyValue,omitempty"`
}

func (m *MetricDesc) String() string {
	return m.Name
}

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
	var variableLabels []string
	for i := range m.VariableLabels {
		variableLabels = append(variableLabels, strings.ReplaceAll(m.VariableLabels[i], ".", "_"))
	}
	variableLabels = append(variableLabels, labels...)

	return prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystem, m.Name),
		m.Help, variableLabels, m.ConstLabels,
	)
}
