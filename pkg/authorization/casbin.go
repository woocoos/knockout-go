package authorization

import (
	"context"
	casbinerr "github.com/casbin/casbin/v2/errors"
	"github.com/gin-gonic/gin"
	"github.com/tsingsun/woocoo/pkg/authz"
	"github.com/tsingsun/woocoo/pkg/conf"
	"github.com/tsingsun/woocoo/pkg/security"
	"github.com/woocoos/casbin-ent-adapter"
	"github.com/woocoos/casbin-ent-adapter/ent"
	"github.com/woocoos/knockout-go/pkg/identity"
	"strconv"
	"strings"
)

// SetAuthorization 设置授权器
func SetAuthorization(cnf *conf.Configuration, client *ent.Client, opts ...entadapter.Option) (authorizer *authz.Authorization, err error) {
	adp, err := entadapter.NewAdapterWithClient(client, opts...)
	if err != nil {
		return
	}
	authz.SetAdapter(adp)
	authorizer, err = authz.NewAuthorization(cnf, authz.WithRequestParseFunc(RBACWithDomainRequestParserFunc))
	if err != nil {
		return
	}
	authz.SetDefaultAuthorization(authorizer)
	return
}

// RBACWithDomainRequestParserFunc 以RBAC with domain模型生成casbin请求
//
// ctx: 一般就是gin.Context
func RBACWithDomainRequestParserFunc(ctx context.Context, id security.Identity, item *security.PermissionItem) []any {
	gctx := ctx.Value(gin.ContextKey).(*gin.Context)
	domain := gctx.GetHeader(identity.TenantHeaderKey)
	p := item.AppCode + ":" + item.Action
	return []any{id.Name(), domain, p, item.Operator}
}

func GetAllowedObjectConditions(user string, action string, prefix string, domain string) ([]string, error) {
	permissions := authz.DefaultAuthorization.BaseEnforcer().GetPermissionsForUserInDomain(user, domain)
	var objectConditions []string
	for _, policy := range permissions {
		// policy {sub, domain, obj, act}
		if policy[3] == action {
			if !strings.HasPrefix(policy[2], prefix) {
				return nil, casbinerr.ErrObjCondition
			}
			objectConditions = append(objectConditions, strings.TrimPrefix(policy[2], prefix))
		}
	}

	if len(objectConditions) == 0 {
		return nil, casbinerr.ErrEmptyCondition
	}

	return objectConditions, nil
}

const (
	ArnSplit   = ":"
	blockSplit = "/"
)

// FormatArnPrefix 资源格式化前缀
func FormatArnPrefix(app, domain, resource string) string {
	return strings.Join([]string{app, domain, resource, ""}, ArnSplit)
}

// ReplaceTenantID 替换资源中的tenant_id
func ReplaceTenantID(input string, tenantID int) string {
	return strings.Replace(input, ArnSplit+"tenant_id"+ArnSplit, ArnSplit+strconv.Itoa(tenantID)+ArnSplit, -1)
}

// FormatResourceArn 格式化资源ARN
func FormatResourceArn(resource string) string {
	subs := strings.Split(resource, ArnSplit)
	var ns []string
	for _, s := range subs {
		if !strings.Contains(s, blockSplit) {
			ns = append(ns, s)
		}
		ss := strings.Split(s, blockSplit)
		if len(ss) > 1 {
			if ss[1] != "" && ss[1] != "*" {
				ns = append(ns, s)
			}
		}
	}
	return strings.Join(ns, ArnSplit)
}
