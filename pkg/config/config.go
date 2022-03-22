package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/creasty/defaults"
	"sigs.k8s.io/yaml"

	"git.irootech.com/sre/queryexporter/pkg/types"
)

type Metric struct {
	*types.Metric `yaml:",inline"`
	DataSources   []*types.DataSource `yaml:"datasources"`
}

type Config struct {
	LogLevel      string               `yaml:"logLevel" default:"info"`
	LogFormat     string               `yaml:"logFormat" default:"console"`
	Addr          string               `yaml:"addr" default:":9696"`
	EnableProfile *bool                `yaml:"enableProfile" default:"true"`
	Servers       []*types.Server      `yaml:"servers"`
	Metrics       map[string][]*Metric `yaml:"metrics"`
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

func ReadFromFile(fn string) (*Config, error) {
	content, err := ioutil.ReadFile(fn)
	if err != nil {
		return nil, err
	}
	data := os.ExpandEnv(string(content))
	var cfg Config
	if err = yaml.Unmarshal([]byte(data), &cfg); err != nil {
		return nil, err
	}
	if err = cfg.validateAndSetDefaults(); err != nil {
		return nil, err
	}
	return &cfg, nil
}
