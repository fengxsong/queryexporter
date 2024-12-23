package types

import "strings"

type DataSourceType string

type Server struct {
	Name string `json:"name"`
	URI  string `json:"uri"`
}

func (s Server) String() string {
	return s.Name
}

type Servers []*Server

type DataSource struct {
	Server   `json:",inline"`
	Database string `json:"database"`
	Table    string `json:"table"`
}

func (ds DataSource) String() string {
	return ds.Server.String()
}

type DataSources []*DataSource

func (dss DataSources) String() string {
	r := make([]string, len(dss))
	for i := range dss {
		r[i] = dss[i].String()
	}
	return strings.Join(r, ",")
}

type Metric struct {
	*MetricDesc `json:",inline"`
	DataSources DataSources `json:"datasources"`
}

type Metrics []*Metric

func (ms Metrics) IterFn(f func(m *Metric) error) error {
	for i := range ms {
		if err := f(ms[i]); err != nil {
			return err
		}
	}
	return nil
}
