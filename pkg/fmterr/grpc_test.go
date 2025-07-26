// grpc_test.go
package fmterr

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/99designs/gqlgen/graphql"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tsingsun/woocoo/contrib/gql"
	"github.com/tsingsun/woocoo/web"
	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tsingsun/woocoo/pkg/conf"
	"github.com/tsingsun/woocoo/web/handler"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	StringStringFloat = "string:%q,string:%q,float:%f"
)

func TestWrapperGrpcStatus(t *testing.T) {
	// 初始化错误码映射，用于测试
	handler.SetErrorMap(map[uint64]string{
		uint64(codes.NotFound):        "Custom404",
		uint64(codes.InvalidArgument): "Custom400",
	}, nil)

	tests := []struct {
		name     string
		inputErr error
		expected error
	}{
		{
			name:     "not grpc error",
			inputErr: errors.New("native"),
			expected: errors.New("native"),
		},
		{
			name:     "gRPC NotFound",
			inputErr: status.Error(codes.NotFound, "NotFound"),
			expected: status.Error(codes.NotFound, "Custom404"),
		},
		{
			name:     "gRPC InvalidArgument",
			inputErr: status.Error(codes.InvalidArgument, "InvalidArgument"),
			expected: status.Error(codes.InvalidArgument, "Custom400"),
		},
		{
			name:     "gRPC Unknown code",
			inputErr: status.Error(codes.Unknown, "unknown"),
			expected: status.Error(codes.Unknown, "unknown"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WrapperGrpcStatus(tt.inputErr)
			assert.EqualError(t, result, tt.expected.Error())
		})
	}
}

func TestUnaryServerInterceptor(t *testing.T) {
	// 初始化错误码映射
	handler.SetErrorMap(map[uint64]string{
		uint64(codes.NotFound): "404",
		10001:                  "Error10001",
		50000:                  StringStringFloat,
	}, nil)

	cfg := conf.New()
	interceptor := UnaryServerInterceptor(cfg)

	t.Run("OK", func(t *testing.T) {
		_, err := interceptor(context.Background(), "request", &grpc.UnaryServerInfo{},
			func(ctx context.Context, req interface{}) (interface{}, error) {
				return "response", nil
			})
		require.NoError(t, err)
	})

	t.Run("has", func(t *testing.T) {
		_, err := interceptor(context.Background(), "request", &grpc.UnaryServerInfo{},
			func(ctx context.Context, req interface{}) (interface{}, error) {
				return nil, status.Error(codes.NotFound, "not found")
			})
		// 验证错误是否被转换
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.EqualError(t, st.Err(), "rpc error: code = NotFound desc = 404")
	})

	t.Run("error", func(t *testing.T) {
		_, err := interceptor(context.Background(), "request", &grpc.UnaryServerInfo{},
			func(ctx context.Context, req interface{}) (interface{}, error) {
				return nil, Code(10001)
			})
		// 验证错误是否被转换
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.ErrorContains(t, st.Err(), "Error10001")
	})

	t.Run("format error", func(t *testing.T) {
		_, err := interceptor(context.Background(), "request", &grpc.UnaryServerInfo{},
			func(ctx context.Context, req interface{}) (interface{}, error) {
				return nil, Codel(50000, "string,string", "string", 1.1)
			})
		// 验证错误是否被转换
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.ErrorContains(t, st.Err(), `{"error":"string:%q,string:%q,float:%f","meta":["string,string","string",1.1]}`)
	})
}

func TestUnaryClientInterceptor(t *testing.T) {
	// 初始化错误码映射
	handler.SetErrorMap(map[uint64]string{
		uint64(codes.NotFound): "404",
		10001:                  "Error10001",
		50000:                  StringStringFloat,
	}, nil)

	interceptor := UnaryClientInterceptorInGin(conf.New())
	t.Run("simple msg", func(t *testing.T) {
		err := interceptor(context.Background(), "method", nil, nil, nil,
			func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
				return WrapperGrpcStatus(status.Error(10001, "not found"))
			})
		assert.ErrorContains(t, err, "Error10001")
	})
	t.Run("format err", func(t *testing.T) {
		err := interceptor(context.Background(), "method", nil, nil, nil,
			func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
				return WrapperGrpcStatus(Codel(50000, "string,string", "string", 1.1))
			})
		assert.ErrorContains(t, err, "string:%q,string:%q,float:%f")
	})
}

func TestGqlToGrpc(t *testing.T) {
	// 初始化错误码映射
	handler.SetErrorMap(map[uint64]string{
		uint64(codes.NotFound): "404",
		10001:                  "Error10001",
		50000:                  StringStringFloat,
		50001:                  "User {{uid}}",
	}, nil)

	var cfgStr = `
web:
  server:
  engine:
    routerGroups:
    - default:
        middlewares:
        - recovery:
        - errorHandle:
        - graphql:
`

	cfg := conf.NewFromBytes([]byte(cfgStr)).AsGlobal()
	srv := web.New(web.WithConfiguration(cfg.Sub("web")),
		gql.RegisterMiddleware(),
	)
	schema := gqlparser.MustLoadSchema(&ast.Source{Input: `
		type Query {
			hello: Boolean!
		}
		type Mutation {
			name: String!
		}
	`})
	const header = "errorType"
	mock := graphql.ExecutableSchemaMock{
		ComplexityFunc: func(typeName string, fieldName string, childComplexity int, args map[string]any) (int, bool) {
			panic("mock out the Complexity method")
		},
		ExecFunc: func(ctx context.Context) graphql.ResponseHandler {
			opCtx := graphql.GetOperationContext(ctx)
			switch opCtx.Operation.Operation {
			case ast.Query:
				ran := false
				return func(ctx context.Context) *graphql.Response {
					if ran {
						return nil
					}
					ran = true
					gctx, _ := gql.FromIncomingContext(ctx)
					interceptor := UnaryClientInterceptorInGin(conf.New())
					ps := gctx.Request.Header.Get(header)
					switch ps {
					case "grpcErr":
						err := interceptor(context.Background(), "method", nil, nil, nil,
							func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
								return WrapperGrpcStatus(Codel(50000, "string,string", "string", 1.1))
							})
						// mock grpc error return
						graphql.AddError(ctx, err)
					case "grpcMetaErr":
						err := interceptor(context.Background(), "method", nil, nil, nil,
							func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
								return WrapperGrpcStatus(Codef(50001, "uid", "string"))
							})
						// mock grpc error return
						graphql.AddError(ctx, err)
					default:
						graphql.AddError(ctx, &gin.Error{
							Err:  errors.New("gin error"),
							Type: 10000, // custom code
						})
					}
					return &graphql.Response{
						Data: []byte(`null`),
					}
				}
			case ast.Mutation:
				return graphql.OneShot(graphql.ErrorResponse(ctx, "mutations are not supported"))
			case ast.Subscription:
				return graphql.OneShot(graphql.ErrorResponse(ctx, "subscription are not supported"))
			default:
				return graphql.OneShot(graphql.ErrorResponse(ctx, "unsupported GraphQL operation"))
			}
		},
		SchemaFunc: func() *ast.Schema {
			return schema
		},
	}
	_, err := gql.RegisterSchema(srv, &mock)
	require.NoError(t, err)
	var reuqest = func(target, uid string) *http.Request {
		r := httptest.NewRequest("POST", target, bytes.NewReader([]byte(`{"query":"query hello { hello }"}`)))
		r.Header.Set("Content-Type", "application/json")
		return r
	}
	t.Run("default", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := reuqest("/query", "1")
		var rt graphql.Response
		srv.Router().ServeHTTP(w, r)
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &rt), w.Body.String())
		assert.Equal(t, float64(10000), rt.Errors[0].Extensions["code"])
	})
	t.Run("grpc error", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := reuqest("/query", "1")
		r.Header.Add(header, "grpcErr")
		var rt graphql.Response
		srv.Router().ServeHTTP(w, r)
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &rt), w.Body.String())
		assert.Equal(t, []any{"string,string", "string", 1.1}, rt.Errors[0].Extensions["meta"])
	})
	t.Run("grpc meta error", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := reuqest("/query", "1")
		r.Header.Add(header, "grpcMetaErr")
		var rt graphql.Response
		srv.Router().ServeHTTP(w, r)
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &rt), w.Body.String())
		assert.Equal(t, map[string]any{"uid": "string"}, rt.Errors[0].Extensions["meta"])
	})
}
