package types

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	TypeGauge = "gauge"
)

type Server struct {
	Name string `json:"name"`
	URI  string `json:"uri"`
}

type DataSource struct {
	Server   `json:",inline"`
	Database string `json:"database"`
	Table    string `json:"table"`
}

type Metric struct {
	Name string `json:"name"`
	Help string `json:"help"`
	Type string `json:"type" default:"gauge"`
	// for dynamic labels from query results
	VariableLabels []string `json:"variableLabels,omitempty"`
	// for taking value from result
	VariableValue string            `json:"variableValue"`
	ConstLabels   prometheus.Labels `json:"constLabels,omitempty"`
	Query         string            `json:"query"`

	desc *prometheus.Desc
}

func (m *Metric) String() string {
	return m.Name
}

func (m *Metric) setDefaults() {}

func (m *Metric) Validate() error {
	switch m.Type {
	case TypeGauge:
	default:
		return fmt.Errorf("unsupported type %s", m.Type)
	}
	return nil
}

// Desc create singleton desc, labels must include server/database/table
func (m *Metric) Desc(namespace, subsystem string, labels ...string) *prometheus.Desc {
	variableLabels := append(m.VariableLabels, labels...)
	if m.desc == nil {
		m.desc = prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, m.Name),
			m.Help, variableLabels, m.ConstLabels,
		)
	}
	return m.desc
}
