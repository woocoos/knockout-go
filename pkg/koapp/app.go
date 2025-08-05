package koapp

import (
	"context"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"fmt"
	"github.com/XSAM/otelsql"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/tsingsun/woocoo"
	"github.com/tsingsun/woocoo/contrib/telemetry"
	"github.com/tsingsun/woocoo/pkg/cache"
	"github.com/tsingsun/woocoo/pkg/cache/lfu"
	"github.com/tsingsun/woocoo/pkg/cache/redisc"
	"github.com/tsingsun/woocoo/pkg/conf"
	"github.com/tsingsun/woocoo/pkg/log"
	"github.com/tsingsun/woocoo/pkg/store/redisx"
	"github.com/tsingsun/woocoo/pkg/store/sqlx"
	"github.com/woocoos/knockout-go/ent/clientx"
	"github.com/woocoos/knockout-go/pkg/snowflake"
	"go.opentelemetry.io/contrib/propagators/b3"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"os"
	"path/filepath"
)

const (
	otelPathName = "otel"
)

// New initializes the Knockout application, will load common configurations from the configuration file and initialize them.
//
// This function will try to load the following configurations from the configuration file:
//
//	Cache: Configuration for caching, currently supports redis and local.
//	Snowflake: Configuration for generating unique IDs, currently supports snowflake. Note that this configuration is global, if there are multiple application instances, you need to ensure that the configuration of each instance is consistent.
func New(opts ...woocoo.Option) *woocoo.App {
	app := woocoo.New(opts...)
	if !app.AppConfiguration().Exists() {
		wd, _ := os.Getwd()
		panic(fmt.Errorf("configuration file not found: %s", filepath.Join(wd, "etc", "app.yaml")))
	}
	BuildAppComponents(app)
	BuildCacheComponents(app.AppConfiguration())
	return app
}

// BuildAppComponents loads components from the configuration file and initializes them. Application level is generally singleton components
func BuildAppComponents(app *woocoo.App) {
	if app.AppConfiguration().IsSet("snowflake") {
		if err := snowflake.SetDefaultNode(app.AppConfiguration().Sub("snowflake")); err != nil {
			panic(err)
		}
	}
	if app.AppConfiguration().IsSet(otelPathName) {
		otelcfg := telemetry.NewConfig(app.AppConfiguration().Sub(otelPathName),
			telemetry.WithPropagator(b3.New()),
		)
		app.RegisterServer(otelServer{otelcfg})
	}
}

// BuildCacheComponents loads cache service components from the configuration file.
func BuildCacheComponents(cnf *conf.AppConfiguration) {
	cnf.Map("cache", func(root string, sub *conf.Configuration) {
		if !sub.IsSet("driverName") {
			return
		}
		var err error
		name := sub.String("driverName")
		if _, err = cache.GetCache(name); err == nil {
			log.Warn(fmt.Errorf("driver already registered for name %q", name))
			return
		}
		switch root {
		case "redis":
			if cnf.IsSet(otelPathName) {
				remote, err := redisx.NewClient(sub)
				if err != nil {
					panic(err)
				}
				// Enable tracing instrumentation.
				if err = redisotel.InstrumentTracing(remote.UniversalClient); err != nil {
					panic(err)
				}
				// Enable metrics instrumentation.
				if err = redisotel.InstrumentMetrics(remote.UniversalClient); err != nil {
					panic(err)
				}
				if _, err = redisc.New(sub, redisc.WithRedisClient(remote)); err != nil {
					panic(err)
				}
			} else {
				if _, err = redisc.New(sub); err != nil {
					panic(err)
				}
			}
		case "local":
			if _, err = lfu.NewTinyLFU(sub); err != nil {
				panic(err)
			}
		}
	})
}

// BuildEntComponents loads ent service components from the configuration file.
func BuildEntComponents(cnf *conf.AppConfiguration) map[string]dialect.Driver {
	vals := make(map[string]dialect.Driver)
	cnf.Map("store", func(root string, sub *conf.Configuration) {
		var (
			err error
			drv dialect.Driver
		)
		driverName := sub.String("driverName")
		// Try to register otel, if there is otel configuration in the configuration, then register.
		if cnf.IsSet(otelPathName) {
			// Register the otelsql wrapper for the provided postgres driver.
			driverName, err = otelsql.Register(driverName,
				otelsql.WithAttributes(semconv.DBSystemKey.String(driverName)),
				otelsql.WithAttributes(semconv.DBNameKey.String(root)),
				otelsql.WithSpanOptions(otelsql.SpanOptions{
					DisableErrSkip:  true,
					OmitRows:        true,
					OmitConnPrepare: true,
				}),
			)
			if err != nil {
				panic(err)
			}
			sub.Parser().Set("driverName", driverName)
		}
		drv = sql.OpenDB(driverName, sqlx.NewSqlDB(sub))
		if cnf.IsSet("entcache") {
			ccnf := cnf.Sub("entcache")
			if cnf.Bool("entcache.isolate") {
				ccnf.Parser().Set("name", root)
			}
			drv, _ = clientx.BuildEntCacheDriver(ccnf, drv)
		}
		vals[root] = drv
	})
	return vals
}

type otelServer struct {
	*telemetry.Config
}

func (o otelServer) Start(ctx context.Context) error {
	return nil
}

func (o otelServer) Stop(ctx context.Context) error {
	o.Shutdown()
	return nil
}
