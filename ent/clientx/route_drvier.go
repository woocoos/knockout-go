package clientx

import (
	"context"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"github.com/tsingsun/woocoo/pkg/conf"
	"github.com/tsingsun/woocoo/pkg/store/sqlx"
	"github.com/woocoos/knockout-go/pkg/identity"
	"strconv"
)

var _ dialect.Driver = (*RouteDriver)(nil)

type RouteDriverOption func(*RouteDriver)

// RouteDriver is a dialect.Driver that routes to different database instances based on the domain name.
type RouteDriver struct {
	// contains filtered or unexported fields
	dbRules       map[string]dialect.Driver
	defaultDriver dialect.Driver
	useDefault    bool
}

// NewRouteDriver return a router driver from woocoo configuration
//
// Example:
//
// ``` yaml
// store:
//
//	portal:
//	  driverName: mysql
//	  dsn: root:@tcp(localhost:3306)/portal?parseTime=true&loc=Local
//	  multiInstances:
//	    test.com:
//	      dsn: root:@tcp(localhost:3307)/portal?parseTime=true&loc=Local
//	    test.cn:
//	      driverName: mysql
//	      dsn: root:@tcp(localhost:3308)/portal?parseTime=true&loc=Local
//
// ```
func NewRouteDriver(cfg *conf.Configuration, opts ...RouteDriverOption) *RouteDriver {
	rd := &RouteDriver{dbRules: make(map[string]dialect.Driver)}
	for _, opt := range opts {
		opt(rd)
	}
	if cfg.IsSet("multiInstances") {
		for domain := range cfg.Sub("multiInstances").AllSettings() {
			subCfg := cfg.Sub("multiInstances." + domain)
			if subCfg.IsSet("driverName") {
				subCfg.Parser().Set("driverName", cfg.String("driverName"))
			}
			db := sqlx.NewSqlDB(subCfg)
			rd.dbRules[domain] = sql.OpenDB(cfg.String("driverName"), db)
		}
	}
	if len(rd.dbRules) == 0 {
		rd.useDefault = true
	}
	if rd.defaultDriver == nil {
		df := sqlx.NewSqlDB(cfg)
		rd.defaultDriver = sql.OpenDB(cfg.String("driverName"), df)
	}
	return rd
}

func (r *RouteDriver) fromContext(ctx context.Context) dialect.Driver {
	if r.useDefault {
		return r.defaultDriver
	}
	// find domain from context
	tid, ok := identity.TenantIDLoadFromContext(ctx)
	if !ok {
		panic(identity.ErrMisTenantID)
	}
	return r.dbRules[strconv.Itoa(tid)]
}

func (r *RouteDriver) Exec(ctx context.Context, query string, args, v any) error {
	return r.fromContext(ctx).Exec(ctx, query, args, v)
}

func (r *RouteDriver) Query(ctx context.Context, query string, args, v any) error {
	return r.fromContext(ctx).Query(ctx, query, args, v)
}

func (r *RouteDriver) Tx(ctx context.Context) (dialect.Tx, error) {
	return r.fromContext(ctx).Tx(ctx)
}

func (r *RouteDriver) BeginTx(ctx context.Context, opts *sql.TxOptions) (dialect.Tx, error) {
	return r.fromContext(ctx).(interface {
		BeginTx(context.Context, *sql.TxOptions) (dialect.Tx, error)
	}).BeginTx(ctx, opts)
}

func (r *RouteDriver) Close() error {
	for _, d := range r.dbRules {
		d.Close()
	}
	return r.defaultDriver.Close()
}

func (r *RouteDriver) Dialect() string {
	return r.defaultDriver.Dialect()
}
