package types

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	TypeGauge = "gauge"
)

type Server struct {
	Name string `yaml:"name"`
	URI  string `yaml:"uri"`
}

type DataSource struct {
	Server `yaml:",inline"`
	Table  string `yaml:"table"`
}

type Metric struct {
	Name string `yaml:"name"`
	Help string `yaml:"help"`
	Type string `yaml:"type" default:"gauge"`
	// for dynamic labels from query results
	VariableLabels []string `yaml:"variableLabels,omitempty"`
	// for taking value from result
	// can be string or $1/$2 index
	VariableValue interface{}       `yaml:"variableValue"`
	ConstLabels   prometheus.Labels `yaml:"constLabels,omitempty"`
	DataSources   []*DataSource     `yaml:"datasources"`
	Query         string            `yaml:"query"`

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
