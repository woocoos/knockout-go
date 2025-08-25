package koapp

import (
	"testing"

	"entgo.io/ent/dialect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tsingsun/woocoo"
	"github.com/tsingsun/woocoo/pkg/cache"
	"github.com/tsingsun/woocoo/pkg/conf"
	"github.com/woocoos/entcache"
	"github.com/woocoos/knockout-go/test"

	_ "github.com/go-sql-driver/mysql"
)

func TestNew(t *testing.T) {
	cnfAll := conf.New(conf.WithBaseDir(test.Path("testdata")),
		conf.WithLocalPath(test.Path("testdata/app/app.yaml"))).Load()
	t.Run("cache", func(t *testing.T) {
		New(woocoo.WithAppConfiguration(cnfAll.Sub("withCache")))
		_, err := cache.GetCache("redis")
		assert.NoError(t, err)
		_, err = cache.GetCache("local")
		assert.Error(t, err, "local has not register as global cache")
	})
	t.Run("caceh-otel", func(t *testing.T) {
		ccfg := cnfAll.Sub("withCache")
		ccfg.Parser().Set(otelPathName, map[string]any{
			"traceExporter": "stdout",
		})
		New(woocoo.WithAppConfiguration(ccfg))
		_, err := cache.GetCache("redis")
		assert.NoError(t, err)
	})
}

func TestBuildEntComponents(t *testing.T) {
	type args struct {
		cnf *conf.AppConfiguration
	}
	tests := []struct {
		name  string
		args  args
		check func(driver map[string]dialect.Driver)
	}{
		{
			name: "otel",
			args: args{
				cnf: &conf.AppConfiguration{
					Configuration: conf.NewFromStringMap(map[string]any{
						"otel": "",
						"store": map[string]any{
							"mysql": map[string]any{
								"driverName": "mysql",
								"dsn":        "root:@tcp(localhost:3306)/portal?parseTime=true&loc=Local",
							},
						},
					}),
				},
			},
			check: func(driver map[string]dialect.Driver) {
				require.Len(t, driver, 1)
				assert.NotNil(t, driver["mysql"])
			},
		},
		{
			name: "entcache-isolate",
			args: args{
				cnf: &conf.AppConfiguration{
					Configuration: conf.NewFromStringMap(map[string]any{
						"store": map[string]any{
							"isolate": map[string]any{
								"driverName": "mysql",
								"dsn":        "root:@tcp(localhost:3306)/portal?parseTime=true&loc=Local",
							},
						},
						"entcache": map[string]any{
							"isolate":      true,
							"hashQueryTTL": "10m",
						},
					}),
				},
			},
			check: func(driver map[string]dialect.Driver) {
				require.Len(t, driver, 1)
				require.NotNil(t, driver["isolate"])
				d := driver["isolate"]
				ed := d.(*entcache.Driver)
				assert.Equal(t, "isolate", ed.Config.Name)
			},
		},
		{
			name: "empty driver",
			args: args{
				cnf: &conf.AppConfiguration{
					Configuration: conf.NewFromStringMap(map[string]any{
						"store": map[string]any{
							"redis": map[string]any{
								"addrs": []string{"localhost:6379"},
							},
						},
					}),
				},
			},
			check: func(driver map[string]dialect.Driver) {
				assert.Len(t, driver, 0)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildEntComponents(tt.args.cnf)
			tt.check(got)
		})
	}
}
