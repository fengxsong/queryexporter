package sql

import (
	"context"
	"database/sql"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"go.uber.org/multierr"

	"github.com/fengxsong/queryexporter/pkg/querier/factory"
	"github.com/fengxsong/queryexporter/pkg/types"
)

type sqlDriver struct {
	driverName string
	cached     sync.Map
}

func (d *sqlDriver) getCachedClient(uri string) (*sql.DB, error) {
	val, ok := d.cached.Load(uri)
	if ok {
		return val.(*sql.DB), nil
	}

	db, err := sql.Open(d.driverName, uri)
	if err != nil {
		return nil, err
	}
	d.cached.Store(uri, db)
	return db, nil
}

func (d *sqlDriver) Query(ctx context.Context, ds *types.DataSource, query string) ([]types.Result, error) {
	db, err := d.getCachedClient(ds.URI)
	if err != nil {
		return nil, err
	}
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	cols, err := rows.Columns() // just ignore error?
	if err != nil {
		return nil, err
	}
	var (
		rets = make([]types.Result, 0)
		errs = make([]error, 0)
	)
	for rows.Next() {
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i, _ := range columns {
			columnPointers[i] = &columns[i]
		}

		// Scan the result into the column pointers...
		if err := rows.Scan(columnPointers...); err != nil {
			// OR do fast return?
			errs = append(errs, err)
			continue
		}

		// Create our map, and retrieve the value for each column from the pointers slice,
		// storing it in the map with the name of the column as the key.
		m := make(types.Result)
		for i, colName := range cols {
			val := columnPointers[i].(*interface{})
			m[colName] = *val
		}
		rets = append(rets, m)
	}
	if len(errs) > 0 {
		return nil, multierr.Combine(errs...)
	}
	return rets, nil
}

func init() {
	factory.Register("mysql", &sqlDriver{driverName: "mysql"})
	factory.Register("postgres", &sqlDriver{driverName: "postgres"})
}
