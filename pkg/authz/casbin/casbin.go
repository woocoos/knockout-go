package casbin

import (
	"context"
	"fmt"
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
	fileadapter "github.com/casbin/casbin/v2/persist/file-adapter"
	rediswatcher "github.com/casbin/redis-watcher/v2"
	"github.com/tsingsun/woocoo/pkg/conf"
	"github.com/tsingsun/woocoo/pkg/security"
	entadapter "github.com/woocoos/casbin-ent-adapter"
	"github.com/woocoos/casbin-ent-adapter/ent"
	"github.com/woocoos/knockout-go/pkg/authz"
	"github.com/woocoos/knockout-go/pkg/identity"
	"strconv"
	"strings"
)

var (
	_              security.Authorizer = (*Authorizer)(nil)
	defaultAdapter persist.Adapter
)

type (
	Option func(*Authorizer)
	// Authorizer is an Authorizer feature base on casbin.
	Authorizer struct {
		Enforcer     casbin.IEnforcer
		baseEnforcer *casbin.Enforcer
		Watcher      persist.Watcher
		autoSave     bool
	}
)

// NewAuthorizer returns a new authenticator with CachedEnforcer and redis watcher by application configuration.
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
//
// .
// autoSave in watcher callback should be false. but set false will cause casbin main nodes lost save data.
// we will improve in the future.current use database unique index to avoid duplicate data.
func NewAuthorizer(cnf *conf.Configuration, opts ...Option) (au *Authorizer, err error) {
	au = &Authorizer{}
	for _, opt := range opts {
		opt(au)
	}
	// model
	var dsl, policy any
	m := cnf.String("model")
	if strings.ContainsRune(m, '\n') {
		dsl, err = model.NewModelFromString(m)
		if err != nil {
			return
		}
	} else {
		dsl = cnf.Abs(cnf.String("model"))
	}
	// policy
	if pv := cnf.String("policy"); pv != "" {
		SetAdapter(fileadapter.NewAdapter(pv))
	}
	policy = defaultAdapter
	enforcer, err := casbin.NewCachedEnforcer(dsl, policy)
	if err != nil {
		return
	}

	if cnf.IsSet("expireTime") {
		enforcer.SetExpireTime(cnf.Duration("expireTime"))
	}
	// autosave default to false, because we use redis watcher
	if cnf.IsSet("autoSave") {
		au.autoSave = cnf.Bool("autoSave")
	}
	enforcer.EnableAutoSave(au.autoSave)

	au.Enforcer = enforcer
	au.baseEnforcer = enforcer.Enforcer
	err = au.buildWatcher(cnf)
	if err != nil {
		return
	}

	return
}

func (au *Authorizer) Prepare(ctx context.Context, kind security.ArnKind, arnParts ...string) (*security.EvalArgs, error) {
	user, ok := security.FromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("security.IsAllow: user not found in context")
	}
	args := &security.EvalArgs{
		User:       user,
		ActionVerb: "read",
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
	return au.Enforcer.Enforce(args.User.Identity().Name(), tenant, string(args.Action), args.ActionVerb)
}

// QueryAllowedResourceConditions returns the allowed resource conditions for the user in domain.
// if the user don't have any permission, return nil.
func (au *Authorizer) QueryAllowedResourceConditions(ctx context.Context, args *security.EvalArgs) ([]string, error) {
	tenant, ok := identity.TenantIDLoadFromContext(ctx)
	if !ok {
		return nil, identity.ErrMisTenantID
	}
	permissions := au.baseEnforcer.GetPermissionsForUserInDomain(args.User.Identity().Name(), strconv.Itoa(tenant))
	if len(permissions) == 0 {
		return nil, nil
	}
	var objectConditions []string
	prefix := string(args.Resource)
	for _, policy := range permissions {
		// policy {sub, domain, obj, act}
		if policy[3] == "read" {
			if !strings.HasPrefix(policy[2], prefix) {
				continue
			}
			objectConditions = append(objectConditions, strings.TrimPrefix(policy[2], prefix))
		}
	}

	return objectConditions, nil
}

func (au *Authorizer) buildWatcher(cnf *conf.Configuration) (err error) {
	if !cnf.IsSet("watcherOptions") {
		return
	}
	watcherOptions := rediswatcher.WatcherOptions{
		OptionalUpdateCallback: rediswatcher.DefaultUpdateCallback(au.Enforcer),
	}
	err = cnf.Sub("watcherOptions").Unmarshal(&watcherOptions)
	if err != nil {
		return
	}

	if watcherOptions.Options.Addr != "" {
		au.Watcher, err = rediswatcher.NewWatcher(watcherOptions.Options.Addr, watcherOptions)
	} else if watcherOptions.ClusterOptions.Addrs != nil {
		au.Watcher, err = rediswatcher.NewWatcherWithCluster(watcherOptions.Options.Addr, watcherOptions)
	}
	if err != nil {
		return
	}
	return au.Enforcer.SetWatcher(au.Watcher)
}

// BaseEnforcer returns the base enforcer. casbin api is not broadcasting to enforcer interface. so need to use base enforcer.
func (au *Authorizer) BaseEnforcer() *casbin.Enforcer {
	return au.baseEnforcer
}

// SetAdapter sets the default adapter for the enforcer.
func SetAdapter(adapter persist.Adapter) {
	defaultAdapter = adapter
}

// SetAuthorizer set the default authorizer for security package.
func SetAuthorizer(cnf *conf.Configuration, client *ent.Client, opts ...entadapter.Option) error {
	adp, err := entadapter.NewAdapterWithClient(client, opts...)
	if err != nil {
		return err
	}
	SetAdapter(adp)
	authorizer, err := NewAuthorizer(cnf)
	if err != nil {
		return err
	}
	security.SetDefaultAuthorizer(authorizer)
	return nil
}
