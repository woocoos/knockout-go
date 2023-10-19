package schemax

import (
	"context"
	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
)

// The Query interface represents an operation that queries with WhereP.
type (
	Query interface {
		WhereP(...func(*sql.Selector))
	}

	Mutator interface {
		Mutate(ctx context.Context, m ent.Mutation) (ent.Value, error)
	}
)
