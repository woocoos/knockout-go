package koapp

import (
	"context"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"fmt"
	"github.com/XSAM/otelsql"
	"github.com/tsingsun/woocoo"
	"github.com/tsingsun/woocoo/contrib/telemetry"
	"github.com/tsingsun/woocoo/pkg/cache"
	"github.com/tsingsun/woocoo/pkg/cache/lfu"
	"github.com/tsingsun/woocoo/pkg/cache/redisc"
	"github.com/tsingsun/woocoo/pkg/conf"
	"github.com/tsingsun/woocoo/pkg/log"
	"github.com/tsingsun/woocoo/pkg/store/sqlx"
	"github.com/woocoos/knockout-go/ent/clientx"
	"github.com/woocoos/knockout-go/pkg/snowflake"
	"go.opentelemetry.io/contrib/propagators/b3"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

const (
	otelPathName = "otel"
)

// New 初始化Knockout应用,尝试从配置文件中加载常用的配置并初始化.
//
// 该函数会尝试从配置文件中加载以下配置:
//
//	Cache: 用于缓存的配置,目前支持redis和local.
//	Snowflake: 用于生成唯一ID的配置,目前支持snowflake.注意该配置是全局的,如果有多个应用实例,需要保证每个实例的配置一致.
func New(opts ...woocoo.Option) *woocoo.App {
	app := woocoo.New(opts...)
	BuildAppComponents(app)
	BuildCacheComponents(app.AppConfiguration())
	return app
}

// BuildAppComponents 从配置文件中加载组件并初始化.应用级一般为单例的组件
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

// BuildCacheComponents 从配置文件中加载缓存服务组件.
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
			_, err = redisc.New(sub)
		case "local":
			_, err = lfu.NewTinyLFU(sub)
		}
		if err != nil {
			panic(err)
		}
	})
}

// BuildEntComponents 从配置文件中加载ent服务组件.
func BuildEntComponents(cnf *conf.AppConfiguration) map[string]dialect.Driver {
	vals := make(map[string]dialect.Driver)
	cnf.Map("store", func(root string, sub *conf.Configuration) {
		var (
			err error
			drv dialect.Driver
		)
		switch driverName := sub.String("driverName"); driverName {
		case dialect.MySQL:
			// 尝试注册otel,如果配置中有otel配置,则注册.
			if cnf.IsSet(otelPathName) {
				// Register the otelsql wrapper for the provided postgres driver.
				driverName, err = otelsql.Register("mysql",
					otelsql.WithAttributes(semconv.DBSystemMySQL),
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
		case dialect.Postgres, dialect.SQLite, dialect.Gremlin:
			drv = sql.OpenDB(driverName, sqlx.NewSqlDB(sub))
		default:
			return
		}
		if cnf.IsSet("entcache") {
			drv, _ = clientx.BuildEntCacheDriver(cnf.Sub("entcache"), drv)
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
