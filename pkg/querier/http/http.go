package http

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"sigs.k8s.io/yaml"

	"git.irootech.com/sre/queryexporter/pkg/querier/factory"
	"git.irootech.com/sre/queryexporter/pkg/types"
)

const name = "http"

type httpDriver struct {
	client *http.Client
}

// parse query into request
type request struct {
	Method  string        `json:"method"`
	URI     string        `json:"uri"`
	Token   string        `json:"token,omitempty"`
	Body    string        `json:"body,omitempty"`
	Headers http.Header   `json:"headers,omitempty"`
	Timeout time.Duration `json:"timeout"`
}

func (d *httpDriver) Query(ctx context.Context, ds *types.DataSource, query string) ([]types.Result, error) {
	var req request
	if err := yaml.Unmarshal([]byte(query), &req); err != nil {
		return nil, err
	}
	if req.Method == "" {
		req.Method = http.MethodGet
	}
	var rawURL string
	if strings.HasPrefix(req.URI, "http") {
		rawURL = req.URI
	} else if len(ds.URI) > 0 {
		u, err := url.Parse(ds.URI)
		if err != nil {
			return nil, err
		}
		u.Path = path.Join(u.Path, req.URI)
		rawURL = u.String()
	}
	var body io.Reader
	if len(req.Body) > 0 {
		body = bytes.NewBufferString(req.Body)
	}
	r, err := http.NewRequest(req.Method, rawURL, body)
	if err != nil {
		return nil, err
	}
	if len(req.Token) > 0 {
		r.Header.Set("Authorization", req.Token)
	}
	if len(req.Headers) > 0 {
		r.Header = req.Headers.Clone()
	}
	if req.Timeout > 0 {
		ctx, cancel := context.WithTimeout(ctx, req.Timeout)
		defer cancel()
		r = r.WithContext(ctx)
	}
	resp, err := d.client.Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var rets []types.Result
	if err = json.NewDecoder(resp.Body).Decode(&rets); err != nil {
		return nil, err
	}
	return rets, nil
}

func init() {
	factory.Register(name, &httpDriver{http.DefaultClient})
}
