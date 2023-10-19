package clientx

import (
	"entgo.io/ent/dialect"
	"github.com/tsingsun/woocoo/pkg/conf"
	"github.com/woocoos/entcache"
	"sync"
)

var (
	ChangeSet     *entcache.ChangeSet
	changeSetOnce sync.Once
)

// BuildEntCacheDriver 构建ent缓存驱动.这里保证ChangeSet为第一次初始的单例值.
func BuildEntCacheDriver(cnf *conf.Configuration, preDriver dialect.Driver) (dialect.Driver, *entcache.ChangeSet) {
	var cacheOpts []entcache.Option
	cacheOpts = append(cacheOpts, entcache.WithConfiguration(cnf))
	changeSetOnce.Do(func() {
		ChangeSet = entcache.NewChangeSet(cnf.Duration("gcInterval"))
	})
	if ChangeSet != nil {
		cacheOpts = append(cacheOpts, entcache.WithChangeSet(ChangeSet))
	}
	drv := entcache.NewDriver(preDriver, cacheOpts...)
	return drv, ChangeSet
}
