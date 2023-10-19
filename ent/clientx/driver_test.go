package clientx

import (
	"entgo.io/ent/dialect"
	"github.com/stretchr/testify/assert"
	"github.com/tsingsun/woocoo/pkg/conf"
	"testing"
)

func TestBuildEntCacheDriver(t *testing.T) {
	t.Run("no set", func(t *testing.T) {
		cnf := &conf.AppConfiguration{
			Configuration: conf.New(),
		}
		var preDriver dialect.Driver
		got, _ := BuildEntCacheDriver(cnf.Configuration, preDriver)
		assert.NotNil(t, got)
	})
}
