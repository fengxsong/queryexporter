package config

import (
	"fmt"
	"io"
	"os"

	"github.com/a8m/envsubst"
	"github.com/creasty/defaults"
	"sigs.k8s.io/yaml"

	"github.com/fengxsong/queryexporter/pkg/types"
)

type Config struct {
	Servers      types.Servers                          `json:"servers"`
	Aggregations map[types.DataSourceType]types.Metrics `json:"aggregations"`
}

func (c *Config) validateAndSetDefaults() error {
	servers := make(map[string]*types.Server, len(c.Servers))
	for _, s := range c.Servers {
		if s.Name == "" {
			return fmt.Errorf("name property is required of uri %s", s.URI)
		}
		if _, ok := servers[s.Name]; ok {
			return fmt.Errorf("duplicate server %s", s.Name)
		}
		servers[s.Name] = s
	}

	if err := defaults.Set(c); err != nil {
		return err
	}

	setf := func(m *types.Metric) error {
		for _, ds := range m.DataSources {
			if _, ok := servers[ds.Name]; !ok {
				return fmt.Errorf("unknown server %s", ds.Name)
			}
			if ds.URI == "" {
				ds.URI = servers[ds.Name].URI
			}
		}
		return m.Validate()
	}

	for _, metrics := range c.Aggregations {
		if err := metrics.IterFn(setf); err != nil {
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

func Dump(c *Config, w io.Writer) {
	out, err := yaml.Marshal(c)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	} else {
		fmt.Fprintf(w, "%s", out)
	}
}
