package casbin

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
	fileadapter "github.com/casbin/casbin/v2/persist/file-adapter"
	stringadapter "github.com/casbin/casbin/v2/persist/string-adapter"
	rediswatcher "github.com/casbin/redis-watcher/v2"
	"github.com/tsingsun/woocoo/pkg/cache"
	"github.com/tsingsun/woocoo/pkg/cache/lfu"
	"github.com/tsingsun/woocoo/pkg/conf"
	"github.com/tsingsun/woocoo/pkg/security"
	"github.com/woocoos/knockout-go/pkg/authz"
	"github.com/woocoos/knockout-go/pkg/identity"
)

var (
	_ security.Authorizer = (*Authorizer)(nil)
)

type (
	Option func(*Authorizer)
	// Authorizer is an Authorizer feature base on casbin.
	Authorizer struct {
		Enforcer casbin.IEnforcer
		Watcher  persist.Watcher
		Adapter  persist.Adapter
		dsl      any
		// AutoSave 一般管理端需要设置为true.
		AutoSave bool `json:"autoSave"`
		// local cache
		cache cache.Cache
	}
)

func WithEnforcer(e casbin.IEnforcer) Option {
	return func(a *Authorizer) {
		a.Enforcer = e
	}
}

func WithWatcher(w persist.Watcher) Option {
	return func(a *Authorizer) {
		a.Watcher = w
	}
}

func WithAdapter(pa persist.Adapter) Option {
	return func(a *Authorizer) {
		a.Adapter = pa
	}
}

// NewAuthorizer 根据配置创建验证器.
// Configuration example:
//
// authz:
//
//	autoSave: false
//	expireTime: 1h
//	watcherOptions:
//	  options:
//	    addr: "localhost:6379"
//	    channel: "/casbin"
//	model: /path/to/model.conf
//	policy: /path/to/policy.csv
//	cache:
//	  size: 1000
//	  ttl:  1m
//
// .
// autoSave in watcher callback should be false. but set false will cause casbin main nodes lost save data.
// we will improve in the future.current use database unique index to avoid duplicate data.
//
// cache.ttl default 1 minute.
func NewAuthorizer(cnf *conf.Configuration, opts ...Option) (au *Authorizer, err error) {
	au = &Authorizer{}
	for _, opt := range opts {
		opt(au)
	}
	if au.Enforcer == nil {
		if au.Adapter == nil {
			pv := cnf.String("policy")
			_, err = os.Stat(pv)
			if err != nil {
				au.Adapter = stringadapter.NewAdapter(pv)
			} else {
				au.Adapter = fileadapter.NewAdapter(pv)
			}
		}

		m := cnf.String("model")
		if strings.ContainsRune(m, '\n') {
			au.dsl, err = model.NewModelFromString(m)
			if err != nil {
				return nil, err
			}
		} else {
			au.dsl = cnf.Abs(cnf.String("model"))
		}
		au.Enforcer, err = casbin.NewEnforcer(au.dsl, au.Adapter)
		if err != nil {
			return nil, err
		}
	}
	if au.AutoSave {
		au.Enforcer.EnableAutoSave(au.AutoSave)
	}
	if au.Watcher == nil && cnf.IsSet("watcherOptions") {
		if err = au.buildRedisWatcher(cnf); err != nil {
			return nil, err
		}
	}
	if cnf.IsSet("cache") {
		if au.cache, err = lfu.NewTinyLFU(cnf.Sub("cache")); err != nil {
			return nil, err
		}
	}

	return au, nil
}

func (au *Authorizer) Prepare(ctx context.Context, kind security.ArnKind, arnParts ...string) (*security.EvalArgs, error) {
	user, ok := security.FromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("security.IsAllow: user not found in context")
	}
	args := &security.EvalArgs{
		User:       user,
		ActionVerb: authz.ActionTypeRead,
	}
	switch kind {
	case security.ArnKindWeb, security.ArnKindGql:
		args.Action = security.Action(arnParts[0] + ":" + arnParts[2])
	case authz.ActionKindResourcePrefix:
		// prefix, add ":" to the end
		args.Resource = security.Resource(strings.Join(arnParts, ":")) + ":"
	default:
		return nil, fmt.Errorf("authz.Prepare not support kind %s", kind)
	}
	return args, nil
}

// Eval checks if the user has permission to do an operation on a resource.
// tenant will be used as domain. tenant allows not set.
func (au *Authorizer) Eval(ctx context.Context, args *security.EvalArgs) (bool, error) {
	tenant, ok := identity.TenantIDLoadFromContext(ctx)
	if !ok {
		return au.Enforcer.Enforce(args.User.Identity().Name(), string(args.Action), args.ActionVerb)
	}
	// read is the access name.
	return au.Enforcer.Enforce(args.User.Identity().Name(), strconv.Itoa(tenant), string(args.Action), args.ActionVerb)
}

// QueryAllowedResourceConditions returns the allowed resource conditions for the user in domain.
// if the user don't have any permission, return nil.
// A ResourceCondition's operation should be use `data`.
func (au *Authorizer) QueryAllowedResourceConditions(ctx context.Context, args *security.EvalArgs) (conditions []string, err error) {
	tenant, ok := identity.TenantIDLoadFromContext(ctx)
	if !ok {
		return nil, identity.ErrMisTenantID
	}
	uid := args.User.Identity().Name()
	tid := strconv.Itoa(tenant)
	prefix := string(args.Resource)
	var cachekey string
	if au.cache != nil {
		cachekey = tid + "_" + uid + "_" + prefix
		if err = au.cache.Get(ctx, cachekey, &conditions); err != nil {
			if !errors.Is(err, cache.ErrCacheMiss) {
				return nil, err
			}
		} else {
			return conditions, nil
		}
	}
	permissions, err := au.Enforcer.GetImplicitPermissionsForUser(args.User.Identity().Name(), tid)
	if err != nil {
		return nil, err
	}
	if len(permissions) > 0 {
		for _, policy := range permissions {
			// policy {sub, domain, obj, act}
			if policy[3] == authz.ActionTypeSchema {
				if !strings.HasPrefix(policy[2], prefix) {
					continue
				}
				conditions = append(conditions, strings.TrimPrefix(policy[2], prefix))
			}
		}
	}
	if au.cache != nil {
		if err = au.cache.Set(ctx, cachekey, conditions); err != nil {
			return nil, err
		}
	}
	return conditions, nil
}

// BaseEnforcer returns the base enforcer. casbin api is not broadcasting to enforcer interface. so need to use base enforcer.
func (au *Authorizer) BaseEnforcer() casbin.IEnforcer {
	return au.Enforcer
}

// SetAuthorizer set the default authorizer for security package.
func SetAuthorizer(cnf *conf.Configuration, opts ...Option) error {
	authorizer, err := NewAuthorizer(cnf, opts...)
	if err != nil {
		return err
	}
	security.SetDefaultAuthorizer(authorizer)
	return nil
}

func (au *Authorizer) buildRedisWatcher(cnf *conf.Configuration) error {
	watcherOptions := rediswatcher.WatcherOptions{
		OptionalUpdateCallback: rediswatcher.DefaultUpdateCallback(au.Enforcer),
	}
	err := cnf.Sub("watcherOptions").Unmarshal(&watcherOptions)
	if err != nil {
		return err
	}
	var watcher persist.Watcher
	if watcherOptions.Options.Addr != "" {
		watcher, err = rediswatcher.NewWatcher(watcherOptions.Options.Addr, watcherOptions)
	} else if watcherOptions.ClusterOptions.Addrs != nil {
		watcher, err = rediswatcher.NewWatcherWithCluster(watcherOptions.Options.Addr, watcherOptions)
	}
	if err != nil {
		return err
	}
	au.Watcher = watcher
	return au.Enforcer.SetWatcher(watcher)
}
