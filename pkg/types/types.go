package types

type DataSourceType string

type Server struct {
	Name string `json:"name"`
	URI  string `json:"uri"`
}

type Servers []*Server

type DataSource struct {
	Server   `json:",inline"`
	Database string `json:"database"`
	Table    string `json:"table"`
}

type DataSources []*DataSource

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
