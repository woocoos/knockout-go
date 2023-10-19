package helloapp

import (
	"context"
	"github.com/golang-jwt/jwt/v4"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tsingsun/woocoo/pkg/authz"
	"github.com/tsingsun/woocoo/pkg/conf"
	"github.com/tsingsun/woocoo/pkg/security"
	entadapter "github.com/woocoos/casbin-ent-adapter"
	casbinent "github.com/woocoos/casbin-ent-adapter/ent"
	"github.com/woocoos/knockout-go/ent/schemax"
	"github.com/woocoos/knockout-go/integration/helloapp/ent"
	_ "github.com/woocoos/knockout-go/integration/helloapp/ent/runtime"
	"github.com/woocoos/knockout-go/integration/helloapp/ent/world"
	"github.com/woocoos/knockout-go/pkg/authorization"
	"github.com/woocoos/knockout-go/pkg/identity"
	"log"
	"math/rand"
	"strconv"
	"testing"
)

func open(ctx context.Context) *ent.Client {
	client, err := ent.Open("sqlite3", "file:ent?mode=memory&cache=shared&_fk=1", ent.Debug())
	if err != nil {
		log.Fatalf("failed opening connection to sqlite: %v", err)
	}
	// Run the auto migration tool.
	if err := client.Schema.Create(ctx); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}
	return client
}

func initCasbin(ctx context.Context) {
	client, err := casbinent.Open("sqlite3", "file:casbin?mode=memory&cache=shared&_fk=1")
	if err != nil {
		log.Fatalf("failed opening connection to sqlite: %v", err)
	}
	_, err = authorization.SetAuthorization(conf.NewFromStringMap(map[string]any{
		"model": `
[request_definition]
r = sub, dom, obj, act
[policy_definition]
p = sub, dom, obj, act, eft
[role_definition]
g = _, _, _
[policy_effect]
e = some(where (p.eft == allow))
[matchers]
m = g(r.sub, p.sub, r.dom) && r.dom == p.dom && keyMatch(r.obj, p.obj) && r.act == p.act
`,
	}), client, entadapter.WithMigration())
	if err != nil {
		log.Fatalf("failed init casbin: %v", err)
	}
}

func Test_CreateWorld(t *testing.T) {
	ctx := context.Background()
	client := open(ctx)
	defer client.Close()

	if err := client.World.Create().Exec(ctx); err == nil {
		t.Fatal("expect tenant creation to fail, but got:", err)
	}

	tctx := identity.WithTenantID(ctx, 1)
	tctx = security.WithContext(tctx, security.NewGenericPrincipalByClaims(jwt.MapClaims{
		"sub": "1",
	}))
	if err := client.World.Create().SetName("woocoo").SetTenantID(1).Exec(tctx); err != nil {
		t.Fatal("expect tenant creation to succeed, but got:", err)
	}

	if _, err := client.World.Query().Count(ctx); err == nil {
		t.Fatal("expect tenant query to fail, but got:", err)
	}

	if _, err := client.World.Query().Count(tctx); err != nil {
		t.Fatal("expect tenant query to succeed, but got:", err)
	}
}

func Test_WorldWithTenant(t *testing.T) {
	ctx := context.Background()
	client := open(ctx)
	defer client.Close()

	initCasbin(ctx)

	// ctx without tenant_id
	if err := client.World.Create().Exec(ctx); err == nil {
		t.Fatal("expect tenant creation to fail, but got:", err)
	}

	tid := rand.Int()
	tctx := identity.WithTenantID(ctx, tid)
	tctx = security.WithContext(tctx, security.NewGenericPrincipalByClaims(map[string]interface{}{"sub": "1"}))
	id := rand.Int()

	r, err := authz.DefaultAuthorization.Enforcer.AddRoleForUserInDomain("1", strconv.Itoa(tid), strconv.Itoa(tid))
	require.NoError(t, err)
	arnp := authorization.FormatArnPrefix("", strconv.Itoa(tid), "World")
	r, err = authz.DefaultAuthorization.Enforcer.AddPolicy("1", arnp, "read", "allow")
	require.NoError(t, err)
	r, err = authz.DefaultAuthorization.Enforcer.AddPolicy("1", strconv.Itoa(tid), arnp+"name/abc", "read", "allow")
	require.NoError(t, err)
	r, err = authz.DefaultAuthorization.Enforcer.AddPolicy("1", strconv.Itoa(tid), arnp+"name/cba:power_by/0", "read", "allow")
	require.NoError(t, err)
	assert.True(t, r)
	// set tenant_id to 1 should be not working
	err = client.World.Create().SetID(id).SetName("abc").SetTenantID(1111).Exec(tctx)
	assert.NoError(t, err)
	assert.False(t, client.World.Query().Where(world.ID(1111)).ExistX(tctx))

	c, err := client.World.Query().Count(tctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, c)

	err = client.World.UpdateOneID(id).SetName("cba").SetPowerBy("0").Exec(tctx)
	assert.NoError(t, err)

	row := client.World.GetX(tctx, id)
	assert.Equal(t, "cba", row.Name)

	assert.NoError(t, client.World.DeleteOneID(id).Exec(tctx))

	err = client.World.Create().SetName("abc").Exec(tctx)
	assert.NoError(t, err)
	c, err = client.World.Query().Where(world.TenantID(tid)).Count(schemax.SkipTenantPrivacy(tctx))
	assert.NoError(t, err)
	assert.Equal(t, 1, c)
}

func Test_SoftDelete(t *testing.T) {
	ctx := context.Background()
	client := open(ctx)
	defer client.Close()

	tid := rand.Int()
	tctx := identity.WithTenantID(ctx, tid)
	tctx = security.WithContext(tctx, security.NewGenericPrincipalByClaims(map[string]interface{}{"sub": "1"}))
	id := rand.Int()
	if err := client.World.Create().SetName("woocoo").SetTenantID(tid).SetID(id).Exec(tctx); err != nil {
		t.Fatal("expect tenant creation to succeed, but got:", err)
	}
	c, err := client.World.Query().Count(tctx)
	require.NoError(t, err)
	assert.Equal(t, 1, c)

	require.NoError(t, client.World.DeleteOneID(id).Exec(tctx))
	c, err = client.World.Query().Count(tctx)
	require.NoError(t, err)
	assert.Equal(t, 0, c)

	tctx = schemax.SkipSoftDelete(tctx)
	c, err = client.World.Query().Count(tctx)
	require.NoError(t, err)
	assert.Equal(t, 1, c)
}
