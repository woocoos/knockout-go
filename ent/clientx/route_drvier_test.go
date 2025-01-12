package clientx

import (
	"github.com/stretchr/testify/assert"
	"github.com/tsingsun/woocoo/pkg/conf"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestNewRouteDriver(t *testing.T) {
	cfg := conf.NewFromStringMap(map[string]any{
		"store": map[string]any{
			"portal": map[string]any{
				"driverName": "sqlite3",
				"dsn":        ":memory:",
				"multiInstances": map[string]any{
					"test.com": map[string]any{
						"driverName": "sqlite3",
						"dsn":        ":memory:",
					},
					"test.cn": map[string]any{
						"driverName": "sqlite3",
						"dsn":        ":memory:",
					},
				},
			},
		},
	})
	type args struct {
		cfg *conf.Configuration
	}
	tests := []struct {
		name  string
		args  args
		check func(driver *RouteDriver)
	}{
		{
			name: "multi",
			args: args{cfg: cfg.Sub("store.portal")},
			check: func(driver *RouteDriver) {
				assert.Equal(t, 2, len(driver.dbRules))
			},
		},
		{
			name: "from-bytes",
			args: args{
				cfg: conf.NewFromBytes([]byte(`
store:
  portal:
    driverName: sqlite3
    dsn: ":memory:"
    multiInstances:
      test.com:
        driverName: sqlite3
        dsn: ":memory:"
`)).Sub("store.portal"),
			},
			check: func(driver *RouteDriver) {
				assert.Equal(t, 1, len(driver.dbRules))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewRouteDriver(tt.args.cfg)
			tt.check(got)
		})
	}
}
