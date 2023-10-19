package schemax

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/mixin"
	"github.com/woocoos/entcache"
)

// NotifyMixin helps to notify when data changed.
type NotifyMixin struct {
	mixin.Schema
}

func (NotifyMixin) Hooks() []ent.Hook {
	return []ent.Hook{
		entcache.DataChangeNotify(),
	}
}
