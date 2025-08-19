package helloapp

import (
	"context"
	"log"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tsingsun/woocoo/pkg/conf"
	"github.com/tsingsun/woocoo/pkg/security"
	entadapter "github.com/woocoos/casbin-ent-adapter"
	casbinent "github.com/woocoos/casbin-ent-adapter/ent"
	"github.com/woocoos/knockout-go/ent/schemax"
	"github.com/woocoos/knockout-go/integration/helloapp/ent"
	"github.com/woocoos/knockout-go/integration/helloapp/ent/domain"
	"github.com/woocoos/knockout-go/integration/helloapp/ent/hello"
	_ "github.com/woocoos/knockout-go/integration/helloapp/ent/runtime"
	"github.com/woocoos/knockout-go/integration/helloapp/ent/world"
	"github.com/woocoos/knockout-go/pkg/authz"
	"github.com/woocoos/knockout-go/pkg/authz/casbin"
	"github.com/woocoos/knockout-go/pkg/identity"
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
	err = casbin.SetAuthorizer(conf.NewFromStringMap(map[string]any{
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

	tn := time.Now()
	err := client.World.Create().SetCreatedAt(tn).Exec(ctx)
	require.Error(t, err, "expect tenant creation to fail, but got nil")

	tctx := identity.WithTenantID(ctx, 1)
	tctx = security.WithContext(tctx, security.NewGenericPrincipalByClaims(jwt.MapClaims{
		"sub": "1",
	}))
	err = client.World.Create().SetName("woocoo").SetTenantID(1).Exec(tctx)
	assert.NoError(t, err, "expect tenant creation to succeed")

	_, err = client.World.Query().Count(ctx)
	assert.Error(t, err, "expect tenant query to fail")

	row, err := client.World.Query().All(tctx)
	require.NoError(t, err, "expect tenant query to succeed")
	assert.EqualValues(t, tn.Unix(), row[0].CreatedAt.Unix())
}

func Test_EntWithTenant(t *testing.T) {
	ctx := context.Background()
	client := open(ctx)
	defer client.Close()

	initCasbin(ctx)

	tid := rand.Int()
	tctx := identity.WithTenantID(ctx, tid)
	tctx = security.WithContext(tctx, security.NewGenericPrincipalByClaims(jwt.MapClaims{"sub": "1"}))
	id := rand.Int()

	authorizer := security.DefaultAuthorizer.(*casbin.Authorizer)
	_, err := authorizer.Enforcer.AddRoleForUserInDomain("1", strconv.Itoa(tid), strconv.Itoa(tid))
	require.NoError(t, err)
	helloArnp := authz.FormatArnPrefix("", strconv.Itoa(tid), "Hello")
	_, err = authorizer.Enforcer.AddPolicy("1", helloArnp, authz.ActionTypeSchema, "allow")
	require.NoError(t, err)
	_, err = authorizer.Enforcer.AddPolicy("1", strconv.Itoa(tid), helloArnp+"name/abc", authz.ActionTypeSchema, "allow")
	require.NoError(t, err)
	// set action policy
	_, err = authorizer.Enforcer.AddPolicy("1", "resource:*", "read", "allow")
	require.NoError(t, err)

	worldArnp := authz.FormatArnPrefix("", strconv.Itoa(tid), "World")
	_, err = authorizer.Enforcer.AddPolicy("1", worldArnp, authz.ActionTypeSchema, "allow")
	require.NoError(t, err)
	_, err = authorizer.Enforcer.AddPolicy("1", strconv.Itoa(tid), worldArnp+"name/abc", authz.ActionTypeSchema, "allow")
	require.NoError(t, err)
	_, err = authorizer.Enforcer.AddPolicy("1", strconv.Itoa(tid), worldArnp+"name/cba:power_by/0", authz.ActionTypeSchema, "allow")
	require.NoError(t, err)

	// set action policy
	_, err = authorizer.Enforcer.AddPolicy("1", "resource:*", "read", "allow")
	require.NoError(t, err)
	// tenant privacy but format wrong(extra ":")
	_, err = authorizer.Enforcer.AddPolicy("1", strconv.Itoa(tid), worldArnp+":name/cba:power_by/0", authz.ActionTypeSchema, "allow")

	require.NoError(t, err)
	t.Run("tenant", func(t *testing.T) {
		// ctx without tenant_id
		if err := client.World.Create().Exec(ctx); err == nil {
			t.Fatal("expect tenant creation to fail, but got:", err)
		}

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
	})
	t.Run("with storageKey", func(t *testing.T) {
		// ctx without tenant_id
		if err := client.Hello.Create().Exec(ctx); err == nil {
			t.Fatal("expect tenant creation to fail, but got:", err)
		}

		// set tenant_id to 1 should be not working
		err = client.Hello.Create().SetID(id).SetName("abc").SetTenantID(1111).Exec(tctx)
		assert.NoError(t, err)
		assert.False(t, client.Hello.Query().Where(hello.ID(1111)).ExistX(tctx))

		c, err := client.Hello.Query().Count(tctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, c)

		err = client.Hello.UpdateOneID(id).SetName("cba").Exec(tctx)
		assert.NoError(t, err)

		_, err = client.Hello.Get(tctx, id)
		assert.Error(t, err, "only query name=abc")

		assert.NoError(t, client.Hello.DeleteOneID(id).Exec(tctx))

		err = client.Hello.Create().SetName("abc").Exec(tctx)
		assert.NoError(t, err)
		c, err = client.Hello.Query().Where(hello.TenantID(tid)).Count(schemax.SkipTenantPrivacy(tctx))
		assert.NoError(t, err)
		assert.Equal(t, 1, c)
	})
	t.Run("multi tenant", func(t *testing.T) {
		worldArnp := authz.FormatArnPrefix("", "111", "World")
		_, err := authorizer.Enforcer.AddRoleForUserInDomain("111", "111", "111")
		require.NoError(t, err)
		_, err = authorizer.Enforcer.AddPolicy("111", strconv.Itoa(111), worldArnp+"tenant_id/[123,345]", authz.ActionTypeSchema, "allow")
		require.NoError(t, err)

		ctx := identity.WithTenantID(ctx, 111)
		ctx = security.WithContext(ctx, security.NewGenericPrincipalByClaims(jwt.MapClaims{"sub": "111"}))
		err = client.World.Create().SetName("abc").SetTenantID(123).Exec(schemax.SkipTenantPrivacy(ctx))
		require.NoError(t, err)
		err = client.World.Create().SetName("bcd").SetTenantID(345).Exec(schemax.SkipTenantPrivacy(ctx))
		require.NoError(t, err)
		c, err := client.World.Query().Count(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 2, c)
	})
	t.Run("tenant with domain", func(t *testing.T) {
		uid := "222"
		arn := authz.FormatArnPrefix("", "900", "Domain")
		_, err := authorizer.Enforcer.AddRoleForUserInDomain(uid, uid, "900")
		require.NoError(t, err)
		// set tenant_id to 1 should be not working
		ctx := identity.WithTenantID(ctx, 900)
		ctx = identity.WithDomainID(ctx, 900)
		ctx = security.WithContext(ctx, security.NewGenericPrincipalByClaims(jwt.MapClaims{"sub": uid}))
		require.NoError(t, client.Domain.Create().SetID(9001).SetName("root").SetDomainID(900).SetTenantID(900).Exec(schemax.SkipTenantPrivacy(ctx)))
		require.NoError(t, client.Domain.Create().SetID(9002).SetName("root").SetDomainID(901).SetTenantID(901).Exec(schemax.SkipTenantPrivacy(ctx)))
		require.NoError(t, client.Domain.Create().SetID(9003).SetName("root").SetDomainID(902).SetTenantID(900).Exec(schemax.SkipTenantPrivacy(ctx)))
		assert.True(t, client.Domain.Query().Where(domain.ID(9001)).ExistX(ctx))
		// SELECT COUNT(`domains`.`id`) FROM `domains` WHERE `domains`.`name` = ? AND `domains`.`domain_id` = ? args=[root 900]
		assert.Equal(t, 1, client.Domain.Query().Where(domain.Name("root")).CountX(ctx))

		arn = authz.FormatArnPrefix("", "901", "Domain")
		uid = "333"
		_, err = authorizer.Enforcer.AddRoleForUserInDomain(uid, uid, "901")
		require.NoError(t, err)
		t.Run("diff domain and tenant", func(t *testing.T) {
			ctx = security.WithContext(identity.WithDomainID(identity.WithTenantID(ctx, 901), 900), security.NewGenericPrincipalByClaims(jwt.MapClaims{"sub": uid}))
			// SELECT COUNT(`domains`.`id`) FROM `domains` WHERE `domains`.`name` = ? AND `domains`.`org_id` = ? args=[root 901]
			assert.Equal(t, 1, client.Domain.Query().Where(domain.Name("root")).CountX(ctx))

			_, err = authorizer.Enforcer.AddPolicy(uid, "901", arn+"name/1", authz.ActionTypeSchema, "allow")
			require.NoError(t, err)
			// SELECT COUNT(`domains`.`id`) FROM `domains` WHERE `domains`.`name` = ? AND `domains`.`org_id` = ? AND `domains`.`name` = ? args=[root 901 1]
			assert.Equal(t, 0, client.Domain.Query().Where(domain.Name("root")).CountX(ctx))

			client.Domain.Create().SetName("1").ExecX(ctx)
			assert.Equal(t, 1, client.Domain.Query().CountX(ctx))
		})

		t.Run("some domain and tenant", func(t *testing.T) {
			ctx = security.WithContext(identity.WithDomainID(identity.WithTenantID(ctx, 901), 901), security.NewGenericPrincipalByClaims(jwt.MapClaims{"sub": uid}))
			// SELECT COUNT(`domains`.`id`) FROM `domains` WHERE `domains`.`name` = ? AND `domains`.`domain_id` = ? AND `domains`.`name` = ? args=[root 901 1]
			assert.Equal(t, 0, client.Domain.Query().Where(domain.Name("root")).CountX(ctx))

			_, err = authorizer.Enforcer.RemovePolicy(uid, "901", arn+"name/1", authz.ActionTypeSchema, "allow")
			require.NoError(t, err)
			client.Domain.Create().SetName("a").ExecX(ctx)
			assert.Equal(t, 1, client.Domain.Query().Where(domain.Name("a")).CountX(ctx))
		})

	})
}

func Test_SoftDelete(t *testing.T) {
	ctx := context.Background()
	client := open(ctx)
	defer client.Close()

	tid := rand.Int()
	tctx := identity.WithTenantID(ctx, tid)
	tctx = security.WithContext(tctx, security.NewGenericPrincipalByClaims(jwt.MapClaims{"sub": "1"}))
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
