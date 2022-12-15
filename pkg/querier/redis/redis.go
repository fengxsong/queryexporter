package redis

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/go-redis/redis/v8"

	"github.com/fengxsong/queryexporter/pkg/querier/factory"
	"github.com/fengxsong/queryexporter/pkg/types"
)

const (
	name         = "redis"
	valueKeyName = "value"
)

type redisDriver struct {
	cached sync.Map
}

func (d *redisDriver) getCachedClient(uri string) (*redis.Client, error) {
	val, ok := d.cached.Load(uri)
	if ok {
		return val.(*redis.Client), nil
	}
	opts, err := redis.ParseURL(uri)
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(opts)
	d.cached.Store(uri, client)
	return client, nil
}

func (d *redisDriver) Query(ctx context.Context, ds *types.DataSource, query string) ([]types.Result, error) {
	client, err := d.getCachedClient(ds.URI)
	if err != nil {
		return nil, err
	}
	return doQuery(ctx, client, query)
}

func doQuery(ctx context.Context, client *redis.Client, query string) ([]types.Result, error) {
	parts := strings.Split(query, " ")
	if len(parts) >= 2 {
		switch strings.ToLower(parts[0]) {
		case "get":
			res, err := client.Get(ctx, parts[1]).Result()
			return []types.Result{{valueKeyName: res}}, err
		case "hget":
			if len(parts) != 3 {
				return nil, fmt.Errorf("unknown hget command: %v", query)
			}
			res, err := client.HGet(ctx, parts[1], parts[2]).Result()
			return []types.Result{{valueKeyName: res}}, err
		case "hgetall":
			res := client.HGetAll(ctx, parts[1])
			if res.Err() != nil {
				return nil, res.Err()
			}
			var ret types.Result
			if err := res.Scan(&ret); err != nil {
				return nil, err
			}
			return []types.Result{ret}, nil
		default:
		}
	}
	return nil, fmt.Errorf("unsupported query %s", query)
}

func init() {
	factory.Register(name, &redisDriver{})
}
