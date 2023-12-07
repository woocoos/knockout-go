package filex

import (
	"context"
	"entgo.io/ent"
	"fmt"
	"github.com/woocoos/knockout-go/api/file"
	"net/url"
	"strings"
)

var (
	sdk *file.FileAPI
)

// TODO
func FileMutationHook(appCode, pathField string) ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			var (
				oldPath string
				newPath string
			)

			value, ok := m.Field(pathField)
			if !ok {
				return next.Mutate(ctx, m)
			}
			newPath = value.(string)
			switch m.Op() {
			case ent.OpCreate:
				p, err := url.Parse(newPath)
				if err != nil {
					return nil, err
				}
				if p.Scheme == "file" {
					if !strings.HasPrefix(p.Path, appCode) {
						return nil, fmt.Errorf("invalid path: %s,must be like:%s/xxx", newPath, appCode)
					}
				}
			case ent.OpUpdateOne, ent.OpUpdate, ent.OpDeleteOne, ent.OpDelete:
				o, err := m.OldField(ctx, pathField)
				if err != nil {
					return nil, err
				}
				oldPath = o.(string)
			}
			v, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}
			// todo: delete old file
			oldPath = strings.TrimPrefix(oldPath, appCode)
			return v, err
		})
	}
}
