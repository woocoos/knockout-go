package koapp

import (
	"entgo.io/ent/dialect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tsingsun/woocoo"
	"github.com/tsingsun/woocoo/pkg/cache"
	"github.com/tsingsun/woocoo/pkg/conf"
	"github.com/woocoos/knockout-go/test"
	"testing"

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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildEntComponents(tt.args.cnf)
			tt.check(got)
		})
	}
}
