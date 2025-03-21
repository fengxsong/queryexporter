package mongo

import (
	"context"
	"sync"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/multierr"

	"github.com/fengxsong/queryexporter/pkg/querier/factory"
	"github.com/fengxsong/queryexporter/pkg/querier/log"
	"github.com/fengxsong/queryexporter/pkg/types"
)

const name = "mongo"

type mongoDriver struct {
	cached sync.Map
}

func (d *mongoDriver) getCachedClient(uri string) (*mongo.Client, error) {
	val, ok := d.cached.Load(uri)
	if ok {
		return val.(*mongo.Client), nil
	}

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	d.cached.Store(uri, client)
	return client, nil
}

func (d *mongoDriver) aggregate(ctx context.Context, uri, db, col string, pipeline bson.A) (*mongo.Cursor, error) {
	client, err := d.getCachedClient(uri)
	if err != nil {
		return nil, err
	}
	return client.Database(db).Collection(col).Aggregate(ctx, pipeline)
}

func (d *mongoDriver) Query(ctx context.Context, ds *types.DataSource, query string) ([]types.Result, error) {
	var pipeline bson.A
	if err := bson.UnmarshalExtJSON([]byte(query), false, &pipeline); err != nil {
		return nil, err
	}
	logger := log.GetLogger(ctx)
	logger.Debug("query", "pipeline", pipeline)

	cur, err := d.aggregate(ctx, ds.URI, ds.Database, ds.Table, pipeline)
	if err != nil {
		return nil, err
	}
	var (
		errs = make([]error, 0)
		rets = make([]types.Result, 0)
	)
	for cur.Next(ctx) {
		ret := make(types.Result)
		err = cur.Decode(&ret)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		rets = append(rets, ret)
	}
	if len(errs) > 0 {
		return nil, multierr.Combine(errs...)
	}
	return rets, nil
}

func (d *mongoDriver) Name() string {
	return name
}

func init() {
	factory.Register(name, &mongoDriver{})
}
