package remote

import (
	"context"
	"testing"
	"time"

	"github.com/casbin/casbin/v2/model"
	stringadapter "github.com/casbin/casbin/v2/persist/string-adapter"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/tsingsun/woocoo"
	"github.com/tsingsun/woocoo/pkg/conf"
	"github.com/tsingsun/woocoo/pkg/security"
	"github.com/tsingsun/woocoo/rpc/grpcx"
	"github.com/woocoos/knockout-go/pkg/authz/casbin"
	"github.com/woocoos/knockout-go/pkg/authz/casbin/proto"
	"github.com/woocoos/knockout-go/pkg/identity"

	casbinv2 "github.com/casbin/casbin/v2"
)

const token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6InRlc3QiLCJzdWIiOiIxIiwiaWF0IjoxOTQyMjY5NDQ0fQ.X4E177vMh2t0whYqL0WuHgU7NqzTuBdKfIXrwA-bKwQ"

type mockCasbinServer struct {
	enforcer *casbinv2.Enforcer
	proto.UnimplementedCasbinServer
}

func newMockCasbinServer(t *testing.T) *mockCasbinServer {
	p := conf.New(conf.WithLocalPath("../testdata/casbin.yaml")).Load()
	m, err := model.NewModelFromString(p.String("remote.model"))
	require.NoError(t, err)
	a := stringadapter.NewAdapter(p.String("remote.policy"))
	e, err := casbinv2.NewEnforcer(m, a)
	require.NoError(t, err)
	return &mockCasbinServer{
		enforcer: e,
	}
}

func (m *mockCasbinServer) Enforce(ctx context.Context, req *proto.EnforceRequest) (*proto.BoolReply, error) {
	params := make([]any, 0, len(req.Params))
	for _, v := range req.Params {
		params = append(params, v)
	}
	res, err := m.enforcer.Enforce(params...)
	if err != nil {
		return nil, err
	}
	return &proto.BoolReply{
		Res: res,
	}, nil
}

func (m *mockCasbinServer) GetImplicitPermissionsForUser(ctx context.Context, req *proto.PermissionRequest) (*proto.Array2DReply, error) {
	res, err := m.enforcer.GetImplicitPermissionsForUser(req.User, req.Domain...)
	if err != nil {
		return nil, err
	}
	return m.wrapPlainPolicy(res), err
}

func (m *mockCasbinServer) wrapPlainPolicy(policy [][]string) *proto.Array2DReply {
	if len(policy) == 0 {
		return &proto.Array2DReply{}
	}

	policyReply := &proto.Array2DReply{}
	policyReply.D2 = make([]*proto.Array2DReplyD, len(policy))
	for e := range policy {
		policyReply.D2[e] = &proto.Array2DReplyD{D1: policy[e]}
	}

	return policyReply
}

type remoteSuite struct {
	suite.Suite

	authorizer   *casbin.Authorizer
	client       *Client
	casbinServer *grpcx.Server
	cfg          *conf.Configuration
	app          *woocoo.App
}

func TestRemote(t *testing.T) {
	suite.Run(t, new(remoteSuite))
}

func (t *remoteSuite) SetupSuite() {
	cnf := conf.New(conf.WithLocalPath("../testdata/remote.yaml")).Load()
	t.app = woocoo.New(woocoo.WithAppConfiguration(cnf))
	authz, err := NewGrpcEnforcer(cnf.Sub("casbinServer.grpc"))
	t.Require().NoError(err)
	authorizer, err := casbin.NewAuthorizer(cnf, casbin.WithEnforcer(authz))
	t.Require().NoError(err)
	t.authorizer = authorizer
	t.casbinServer = grpcx.New(grpcx.WithConfiguration(t.app.AppConfiguration().Sub("casbinServer.grpc")))
	proto.RegisterCasbinServer(t.casbinServer.Engine(), newMockCasbinServer(t.T()))

	t.app.RegisterServer(t.casbinServer)
	go func() {
		if err := t.app.Run(); err != nil {
			t.FailNow("app start error")
		}
	}()
	time.Sleep(time.Second)
}

func (t *remoteSuite) TeardownSuite() {
	t.app.Stop()
}

func (t *remoteSuite) TestEval() {
	t.Run("allow", func() {
		ctx := identity.WithTenantID(context.Background(), 1000)
		user := security.NewGenericPrincipalByClaims(jwt.MapClaims{"sub": "alice"})
		ctx = security.WithContext(ctx, user)
		res, err := t.authorizer.Eval(ctx, &security.EvalArgs{
			User:       user,
			Action:     "data1",
			ActionVerb: "read",
		})
		t.Require().NoError(err)
		t.Require().True(res)
	})
}

func (t *remoteSuite) TestQueryAllowedResourceConditions() {
	t.Run("allow", func() {
		ctx := identity.WithTenantID(context.Background(), 1000)
		user := security.NewGenericPrincipalByClaims(jwt.MapClaims{"sub": "alice"})
		ctx = security.WithContext(ctx, user)
		res, err := t.authorizer.QueryAllowedResourceConditions(ctx, &security.EvalArgs{
			User:       user,
			Resource:   ":1000:World",
			ActionVerb: "allow",
		})
		t.Require().NoError(err)
		t.Len(res, 1)
		t.Equal(":name/cba:power_by/0", res[0])
	})
}
