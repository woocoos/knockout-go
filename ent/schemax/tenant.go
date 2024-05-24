package schemax

import (
	"context"
	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"fmt"
	"github.com/tsingsun/woocoo/pkg/security"
	"github.com/woocoos/knockout-go/pkg/authz"
	"github.com/woocoos/knockout-go/pkg/identity"
	"strconv"
	"strings"
)

var (
	FieldTenantID = "tenant_id"
)

type tenantPrivacyKey struct{}

// SkipTenantPrivacy returns a new context that skips the TenantRule interceptor/mutators.
func SkipTenantPrivacy(parent context.Context) context.Context {
	return context.WithValue(parent, tenantPrivacyKey{}, true)
}

// ifSkipTenantPrivacy returns true if the TenantRule interceptor/mutators should be skipped.
func ifSkipTenantPrivacy(ctx context.Context) bool {
	skip, _ := ctx.Value(tenantPrivacyKey{}).(bool)
	return skip
}

// TenantMixin helps to generate a tenant_id field and inject resource query.
//
//	 type World struct {
//		    ent.Schema
//	 }
//
//	 func (World) Mixin() []ent.Mixin {
//		    return []ent.Mixin{
//		    	schemax.NewTenantMixin[intercept.Query, *gen.Client](intercept.NewQuery),
//		    }
//	 }
type TenantMixin[T Query, Q Mutator] struct {
	mixin.Schema
	// application code, defined in configuration file `appName`
	app string
	// the NewQuery returns the generic Query interface for the given typed query.
	newQueryFunc func(ent.Query) (T, error)
	// schemaType overrides the default database type with a custom
	// schema type (per dialect) for int.
	schemaType map[string]string
}

type TenantMixinOption[T Query, Q Mutator] func(*TenantMixin[T, Q])

// WithTenantMixinSchemaType sets the tenant field for ent SchemaType. By default,
// in NewTenantMixin, it will be set to bigint for SnowFlakeID.
func WithTenantMixinSchemaType[T Query, Q Mutator](schemaType map[string]string) TenantMixinOption[T, Q] {
	return func(m *TenantMixin[T, Q]) {
		m.schemaType = schemaType
	}
}

// NewTenantMixin returns a mixin that adds a tenant_id field and inject resource query.
//
// app is the application code, the same as the one defined in knockout backend.
// Knockout Tenant field uses go Int type as the field type, it is a snowflake id by default.
func NewTenantMixin[T Query, Q Mutator](app string, newQuery func(ent.Query) (T, error), opts ...TenantMixinOption[T, Q]) TenantMixin[T, Q] {
	val := TenantMixin[T, Q]{
		app:          app,
		newQueryFunc: newQuery,
		schemaType:   SnowFlakeID{}.SchemaType(),
	}
	for _, opt := range opts {
		opt(&val)
	}
	return val
}

func (d TenantMixin[T, Q]) Fields() []ent.Field {
	return []ent.Field{
		field.Int(FieldTenantID).Immutable().SchemaType(d.schemaType),
	}
}

// Interceptors of the SoftDeleteMixin.
func (d TenantMixin[T, Q]) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		ent.TraverseFunc(func(ctx context.Context, q ent.Query) error {
			if ifSkipTenantPrivacy(ctx) {
				return nil
			}

			df, err := d.newQueryFunc(q)
			if err != nil {
				return err
			}
			return d.QueryRulesP(ctx, df)
		}),
	}
}

type tenant[Q Mutator] interface {
	Query
	Client() Q
	SetTenantID(int)
}

// Hooks of the SoftDeleteMixin.
func (d TenantMixin[T, Q]) Hooks() []ent.Hook {
	return []ent.Hook{
		func(next ent.Mutator) ent.Mutator {
			return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
				if ifSkipTenantPrivacy(ctx) {
					return next.Mutate(ctx, m)
				}

				tid, ok := identity.TenantIDLoadFromContext(ctx)
				if !ok {
					return nil, identity.ErrMisTenantID
				}

				mx, ok := m.(tenant[Q])
				if !ok {
					return nil, fmt.Errorf("unexpected mutation type %T", m)
				}
				switch m.Op() {
				case ent.OpCreate:
					mx.SetTenantID(tid)
				default:
					d.P(mx, tid)
				}
				return next.Mutate(ctx, m)
			})
		},
	}
}

// P adds a storage-level predicate to the queries and mutations.
func (d TenantMixin[T, Q]) P(w Query, tid int) {
	w.WhereP(
		sql.FieldEQ(FieldTenantID, tid),
	)
}

// QueryRulesP adds a storage-level predicate to the queries.
//
// When call Authorizer.Prepare, pass appcode, tenant id, and resource type those as resource prefix,
// the prefix format is `appcode:tenant_id:resource_type:`.
func (d TenantMixin[T, Q]) QueryRulesP(ctx context.Context, w Query) error {
	tid, ok := identity.TenantIDLoadFromContext(ctx)
	if !ok {
		return identity.ErrMisTenantID
	}
	tidstr := strconv.Itoa(tid)
	authArgs, err := security.DefaultAuthorizer.Prepare(ctx, authz.ActionKindResourcePrefix,
		d.app, tidstr, ent.QueryFromContext(ctx).Type)
	if err != nil {
		return err
	}
	flts, err := security.DefaultAuthorizer.QueryAllowedResourceConditions(ctx, authArgs)
	if err != nil {
		return err
	}
	if len(flts) == 0 {
		d.P(w, tid)
		return nil
	}

	w.WhereP(func(selector *sql.Selector) {
		rules := getTenantRules(flts, tidstr, selector)
		if len(rules) > 0 {
			selector.Where(sql.Or(rules...))
		}
	})
	return nil
}

// getTenantRules returns the tenant resource conditions for the current user.
// if field rule is not having value after "/", it will be ignored, and like * effect.
func getTenantRules(filers []string, tid string, selector *sql.Selector) []*sql.Predicate {
	v := make([]*sql.Predicate, 0, len(filers))
	for _, flt := range filers {
		if flt == "" {
			v = append(v, sql.EQ(selector.C(FieldTenantID), tid))
			continue
		}
		fs := strings.Split(flt, ":")
		fv := make([]*sql.Predicate, 0, len(fs))
		for _, f := range fs {
			kvs := strings.Split(f, "/")
			if len(kvs) != 2 {
				continue
			}
			fv = append(fv, sql.EQ(selector.C(kvs[0]), kvs[1]))
		}
		if len(fv) == 1 {
			v = append(v, fv...)
		} else {
			v = append(v, sql.And(fv...))
		}
	}
	return v
}
