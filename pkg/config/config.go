package config

import (
	"fmt"
	"os"

	"github.com/a8m/envsubst"
	"github.com/creasty/defaults"
	"sigs.k8s.io/yaml"

	"github.com/fengxsong/queryexporter/pkg/types"
)

type Metric struct {
	*types.Metric `json:",inline"`
	DataSources   []*types.DataSource `json:"datasources"`
}

type Config struct {
	LogLevel      string               `json:"logLevel" default:"info"`
	LogFormat     string               `json:"logFormat" default:"console"`
	Addr          string               `json:"addr" default:":9696"`
	EnableProfile *bool                `json:"enableProfile" default:"true"`
	Servers       []*types.Server      `json:"servers"`
	Metrics       map[string][]*Metric `json:"metrics"`
}

func (c *Config) validateAndSetDefaults() error {
	servers := make(map[string]*types.Server, len(c.Servers))
	for _, server := range c.Servers {
		if _, ok := servers[server.Name]; ok {
			return fmt.Errorf("duplicate server %s", server.Name)
		}
		servers[server.Name] = server
	}
	if err := defaults.Set(c); err != nil {
		return err
	}
	setFunc := func(metrics []*Metric) error {
		for i := range metrics {
			m := metrics[i]
			for j := range m.DataSources {
				ds := m.DataSources[j]
				if _, ok := servers[ds.Name]; !ok {
					return fmt.Errorf("unknown server %s", ds.Name)
				}
				if ds.URI == "" {
					ds.URI = servers[ds.Name].URI
				}
			}
			if err := m.Validate(); err != nil {
				return err
			}
		}
		return nil
	}
	for _, metrics := range c.Metrics {
		if err := setFunc(metrics); err != nil {
			return err
		}
	}
	return nil
}

func ReadFromFile(fn string, expandEnv bool) (*Config, error) {
	var (
		data []byte
		err  error
	)
	if expandEnv {
		data, err = envsubst.ReadFile(fn)
	} else {
		data, err = os.ReadFile(fn)
	}
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err = yaml.Unmarshal([]byte(data), &cfg); err != nil {
		return nil, err
	}
	if err = cfg.validateAndSetDefaults(); err != nil {
		return nil, err
	}
	return &cfg, nil
}
