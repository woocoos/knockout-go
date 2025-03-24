package idgrpc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tsingsun/woocoo/pkg/conf"
	"github.com/woocoos/knockout-go/pkg/identity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestExtractTenantID(t *testing.T) {
	handler := &TenantHandler{}
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(identity.TenantHeaderKey, "123"))

	newCtx, err := handler.ExtractTenantID(ctx, identity.TenantHeaderKey)
	assert.NoError(t, err)

	id, ok := identity.TenantIDLoadFromContext(newCtx)
	assert.True(t, ok)
	assert.Equal(t, 123, id)
}

func TestUnaryClientInterceptor(t *testing.T) {
	handler := &TenantHandler{}
	cfg := conf.NewFromStringMap(map[string]any{
		headerKeyPath: "testHeader",
	})
	interceptor := handler.UnaryClientInterceptor(cfg)

	ctx := identity.WithTenantID(context.Background(), 123)
	err := interceptor(ctx, "method", nil, nil, nil, func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		md, ok := metadata.FromOutgoingContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, "123", md.Get("testHeader")[0])
		return nil
	})

	assert.NoError(t, err)
}

func TestStreamClientInterceptor(t *testing.T) {
	handler := &TenantHandler{}
	cfg := conf.New()
	interceptor := handler.StreamClientInterceptor(cfg)

	ctx := identity.WithTenantID(context.Background(), 123)
	_, err := interceptor(ctx, nil, nil, "method", func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		md, ok := metadata.FromOutgoingContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, "123", md.Get(identity.TenantHeaderKey)[0])
		return nil, nil
	})

	assert.NoError(t, err)
}

func TestUnaryServerInterceptor(t *testing.T) {
	handler := &TenantHandler{}
	cfg := conf.New()
	interceptor := handler.UnaryServerInterceptor(cfg)

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(identity.TenantHeaderKey, "123"))
	_, err := interceptor(ctx, nil, nil, func(ctx context.Context, req any) (any, error) {
		id, ok := identity.TenantIDLoadFromContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, 123, id)
		return nil, nil
	})

	assert.NoError(t, err)
}

func TestStreamServerInterceptor(t *testing.T) {
	handler := &TenantHandler{}
	cfg := conf.New()
	interceptor := handler.StreamServerInterceptor(cfg)

	stream := &mockServerStream{ctx: metadata.NewIncomingContext(context.Background(), metadata.Pairs(identity.TenantHeaderKey, "123"))}
	err := interceptor(nil, stream, nil, func(srv any, stream grpc.ServerStream) error {
		id, ok := identity.TenantIDLoadFromContext(stream.Context())
		assert.True(t, ok)
		assert.Equal(t, 123, id)
		return nil
	})

	assert.NoError(t, err)
}

type mockServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (m *mockServerStream) Context() context.Context {
	return m.ctx
}
