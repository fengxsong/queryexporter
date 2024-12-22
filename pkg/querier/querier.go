package querier

import (
	_ "github.com/fengxsong/queryexporter/pkg/querier/http"
	_ "github.com/fengxsong/queryexporter/pkg/querier/mongo"
	_ "github.com/fengxsong/queryexporter/pkg/querier/redis"
	_ "github.com/fengxsong/queryexporter/pkg/querier/sql"
)
