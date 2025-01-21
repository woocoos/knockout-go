package authz

import (
	"github.com/tsingsun/woocoo/pkg/security"
	"strconv"
	"strings"
)

const (
	ActionTypeRead  string = "read"
	ActionTypeWrite string = "write"
	// ActionTypeSchema to resource: ent schema and so on
	ActionTypeSchema string = "schema"
)

const (
	fieldTenantID = "tenant_id"
	ArnSplit      = ":"
	blockSplit    = "/"
)

const (
	ActionKindResourcePrefix security.ArnKind = "resourcePrefix"
)

// FormatArnPrefix 资源格式化前缀
func FormatArnPrefix(app, domain, resource string) string {
	return strings.Join([]string{app, domain, resource, ""}, ArnSplit)
}

// ReplaceTenantID 替换资源中的tenant_id
func ReplaceTenantID(input string, tenantID int) string {
	return strings.Replace(input, ArnSplit+fieldTenantID+ArnSplit, ArnSplit+strconv.Itoa(tenantID)+ArnSplit, -1)
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
