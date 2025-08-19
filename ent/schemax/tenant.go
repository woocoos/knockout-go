package schemax

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/schema/mixin"
	"github.com/tsingsun/woocoo/pkg/security"
	"github.com/woocoos/knockout-go/pkg/authz"
	"github.com/woocoos/knockout-go/pkg/identity"
)

const (
	FieldTenantID = "tenant_id"
)

type tenantPrivacyKey struct{}

// SkipTenantPrivacy returns a new context that skips the TenantRule interceptor/mutators.
func SkipTenantPrivacy(parent context.Context) context.Context {
	return context.WithValue(parent, tenantPrivacyKey{}, true)
}

// IfSkipTenantPrivacy returns true if the TenantRule interceptor/mutators should be skipped.
func IfSkipTenantPrivacy(ctx context.Context) bool {
	skip, _ := ctx.Value(tenantPrivacyKey{}).(bool)
	return skip
}

// TenantMixin helps to generate a tenant_id field and inject resource query.
//
//		 type World struct {
//			    ent.Schema
//		 }
//
//		 func (World) Mixin() []ent.Mixin {
//			    return []ent.Mixin{
//			    	schemax.NewTenantMixin[intercept.Query, *gen.Client](intercept.NewQuery),
//			    }
//		 }
//	  func (World) Fields() []ent.Field {
//				return []ent.Field{
//					field.Int(schemax.FieldTenantID).Immutable(),
//				}
//	  }
type TenantMixin[T Query, Q Mutator] struct {
	mixin.Schema
	// application code, defined in configuration file `appName`
	app string
	// the NewQuery returns the generic Query interface for the given typed query.
	newQueryFunc func(ent.Query) (T, error)
	// tenantStorageKey is the key used to ent StorageKey.
	tenantStorageKey string
	// domainStorageKey is the key used to ent StorageKey for domain, default is empty.
	// domain it the root tenant, which is the top-level tenant in the hierarchy.
	// It is used to distinguish between the root tenant and sub-tenants.
	domainStorageKey string
}

type TenantMixinOption[T Query, Q Mutator] func(*TenantMixin[T, Q])

// WithTenantMixinStorageKey sets the tenant field for ent StorageKey if you custom the field name which is not `tenant_id`.
func WithTenantMixinStorageKey[T Query, Q Mutator](storageKey string) TenantMixinOption[T, Q] {
	return func(m *TenantMixin[T, Q]) {
		m.tenantStorageKey = storageKey
	}
}

// WithTenantMixinDomainStorageKey sets the domain field for ent StorageKey if you custom the field name which is not `domain_id`.
func WithTenantMixinDomainStorageKey[T Query, Q Mutator](storageKey string) TenantMixinOption[T, Q] {
	return func(m *TenantMixin[T, Q]) {
		m.domainStorageKey = storageKey
	}
}

// NewTenantMixin returns a mixin that adds a tenant_id field and inject resource query.
//
// app is the application code, the same as the one defined in knockout backend.
// Knockout Tenant field uses go Int type as the field type, it is a snowflake id by default.
func NewTenantMixin[T Query, Q Mutator](app string, newQuery func(ent.Query) (T, error), opts ...TenantMixinOption[T, Q]) TenantMixin[T, Q] {
	val := TenantMixin[T, Q]{
		app:              app,
		newQueryFunc:     newQuery,
		tenantStorageKey: FieldTenantID,
	}
	for _, opt := range opts {
		opt(&val)
	}
	return val
}

// Interceptors of the TenantMixin.
func (d TenantMixin[T, Q]) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		ent.TraverseFunc(func(ctx context.Context, q ent.Query) error {
			if IfSkipTenantPrivacy(ctx) {
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

type domain[Q Mutator] interface {
	SetDomainID(int)
}

// Hooks of the SoftDeleteMixin.
func (d TenantMixin[T, Q]) Hooks() []ent.Hook {
	return []ent.Hook{
		func(next ent.Mutator) ent.Mutator {
			return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
				if IfSkipTenantPrivacy(ctx) {
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
					if d.domainStorageKey != "" {
						// If domainStorageKey is set, we also set the domain id.
						dm, ok := m.(domain[Q])
						if !ok {
							return nil, fmt.Errorf("unexpected mutation type %T", m)
						}
						did, ok := identity.DomainIDLoadFromContext(ctx)
						if !ok {
							return nil, identity.ErrMissDomainTenantID
						}
						dm.SetDomainID(did)
					}
				default:
					var did int
					if d.domainStorageKey != "" {
						did, ok = identity.DomainIDLoadFromContext(ctx)
						if !ok {
							return nil, identity.ErrMissDomainTenantID
						}
					}
					// for update and delete operation, only add tenant_id predicate.
					// it means the domain user can mutate the tenant data in the domain.
					d.P(mx, tid, did)
				}
				return next.Mutate(ctx, m)
			})
		},
	}
}

// P adds a storage-level predicate to the queries and mutations.
func (d TenantMixin[T, Q]) P(w Query, tenantId, domainId int) {
	if tenantId == domainId {
		// if tenantId equals domainId, it is the root tenant, only filter by domainId.
		w.WhereP(
			sql.FieldEQ(d.domainStorageKey, domainId),
		)
		return
	}
	w.WhereP(
		sql.FieldEQ(d.tenantStorageKey, tenantId),
	)
}

// QueryRulesP adds a storage-level predicate to the queries.
//
// When call Authorizer.Prepare, pass appcode, tenant id, and resource type those as resource prefix,
// the prefix format is `appcode:tenant_id:resource_type:expression`.
// The expression is a list of field and value pairs separated by `/`, and the field and value are separated by `:`.
// It means multiple and conditions that only support the equal operation.
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
	var did int
	if d.domainStorageKey != "" {
		did, ok = identity.DomainIDLoadFromContext(ctx)
		if !ok {
			return identity.ErrMissDomainTenantID
		}
	}
	if len(flts) == 0 {
		d.P(w, tid, did)
		return nil
	}

	w.WhereP(func(selector *sql.Selector) {
		rules := d.getTenantRules(flts, tid, did, selector)
		if len(rules) > 0 {
			selector.Where(sql.Or(rules...))
		}
	})
	return nil
}

// getTenantRules returns the tenant resource conditions for the current user.
// resource expression: ("*" | <resource_string> | [<resource_string>, <resource_string>, ...]).
// tenant id is always added as a condition.the fragments are separated by ":".
// if field rule is not having value after "/", it will be ignored, and like * effect.
func (d TenantMixin[T, Q]) getTenantRules(filers []string, tenantId, domainId int, selector *sql.Selector) []*sql.Predicate {
	v := make([]*sql.Predicate, 0, len(filers))
	useDomain := d.domainStorageKey != "" && tenantId == domainId
	for _, flt := range filers {
		tids := []any{tenantId}
		if flt == "" {
			continue
		}
		fs := strings.Split(flt, ":")
		fv := make([]*sql.Predicate, 0, len(fs))
		for _, f := range fs {
			kvs := strings.Split(f, "/")
			if len(kvs) != 2 {
				continue
			}
			k := kvs[0]
			vs := kvs[1]
			switch vs[0] {
			case '[':
				vt := strings.Split(vs[1:len(vs)-1], ",")
				avt := make([]any, len(vt))
				for i, sv := range vt {
					avt[i] = sv
				}
				if k == d.tenantStorageKey {
					tids = append(tids, avt...)
					continue
				} else {
					fv = append(fv, sql.In(selector.C(k), avt...))
				}
			default:
				if k == d.tenantStorageKey {
					if vs != "*" {
						tids = append(tids, vs)
					}
					// TODO supports subs tenant
					continue
				} else if vs != "*" {
					fv = append(fv, sql.EQ(selector.C(k), vs))
				}
			}
		}
		if len(tids) == 1 {
			if useDomain {
				fv = slices.Insert(fv, 0, sql.EQ(selector.C(d.domainStorageKey), tenantId))
			} else {
				fv = slices.Insert(fv, 0, sql.EQ(selector.C(d.tenantStorageKey), tenantId))
			}
		} else {
			// if there are multiple tenant ids, use IN condition.
			fv = slices.Insert(fv, 0, sql.In(selector.C(d.tenantStorageKey), tids...))
		}
		if l := len(fv); l > 1 {
			v = append(v, sql.And(fv...))
		} else if l == 1 {
			v = append(v, fv[0])
		}
	}
	if len(v) == 0 {
		if useDomain {
			v = append(v, sql.EQ(selector.C(d.domainStorageKey), domainId))
		} else {
			v = append(v, sql.EQ(selector.C(d.tenantStorageKey), tenantId))
		}
	}
	return v
}
