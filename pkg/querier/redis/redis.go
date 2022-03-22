package redis

import (
	"context"

	"git.irootech.com/sre/queryexporter/pkg/querier/factory"
	"git.irootech.com/sre/queryexporter/pkg/types"
)

// just placeholder for now

const name = "redis"

type redisDriver struct {
}

func (d *redisDriver) Query(ctx context.Context, ds *types.DataSource, query string) ([]types.Result, error) {
	return nil, nil
}

func init() {
	factory.Register(name, &redisDriver{})
}
