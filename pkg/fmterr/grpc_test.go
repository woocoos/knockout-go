// grpc_test.go
package fmterr

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/tsingsun/woocoo/pkg/conf"
	"github.com/tsingsun/woocoo/web/handler"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
		50000:                  "string:%q,string:%q,float:%f",
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
