package idgrpc

import (
	"context"
	"github.com/golang-jwt/jwt/v5"
	"github.com/tsingsun/woocoo/pkg/conf"
	"github.com/tsingsun/woocoo/pkg/security"
	"github.com/tsingsun/woocoo/rpc/grpcx"
	"github.com/tsingsun/woocoo/rpc/grpcx/interceptor"
	"github.com/woocoos/knockout-go/pkg/identity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"strconv"
)

const (
	tenantName       = "tenant"
	tidHeaderKeyPath = "headerKey"
	userName         = "user"
)

func init() {
	th := &TenantHandler{}
	grpcx.RegisterUnaryClientInterceptor(tenantName, th.UnaryClientInterceptor)
	grpcx.RegisterStreamClientInterceptor(tenantName, th.StreamClientInterceptor)
	grpcx.RegisterGrpcUnaryInterceptor(tenantName, th.UnaryServerInterceptor)
	grpcx.RegisterGrpcStreamInterceptor(tenantName, th.StreamServerInterceptor)
	id := &IdentityHandler{}
	grpcx.RegisterUnaryClientInterceptor(userName, id.UnaryClientInterceptor)
	grpcx.RegisterStreamClientInterceptor(userName, id.StreamClientInterceptor)
	grpcx.RegisterGrpcUnaryInterceptor(userName, id.UnaryServerInterceptor)
	grpcx.RegisterGrpcStreamInterceptor(userName, id.StreamServerInterceptor)
}

type TenantHandler struct{}

// ExtractTenantID extracts the tenant ID from the metadata and returns the updated context.
func (t *TenantHandler) ExtractTenantID(ctx context.Context, headerKey string) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if ids := md.Get(headerKey); len(ids) > 0 {
			if id, err := strconv.Atoi(ids[0]); err == nil {
				return identity.WithTenantID(ctx, id), nil
			} else {
				return ctx, err
			}
		}
	}
	return ctx, nil
}

func getHeaderKey(cfg *conf.Configuration) string {
	headerKey := cfg.String(tidHeaderKeyPath)
	if headerKey == "" {
		headerKey = identity.TenantHeaderKey
	}
	return headerKey
}

func (t *TenantHandler) UnaryClientInterceptor(cfg *conf.Configuration) grpc.UnaryClientInterceptor {
	headerKey := getHeaderKey(cfg)
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if id, ok := identity.TenantIDLoadFromContext(ctx); ok {
			ctx = metadata.AppendToOutgoingContext(ctx, headerKey, strconv.Itoa(id))
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func (t *TenantHandler) StreamClientInterceptor(cfg *conf.Configuration) grpc.StreamClientInterceptor {
	headerKey := getHeaderKey(cfg)
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		if id, ok := identity.TenantIDLoadFromContext(ctx); ok {
			ctx = metadata.AppendToOutgoingContext(ctx, headerKey, strconv.Itoa(id))
		}
		return streamer(ctx, desc, cc, method, opts...)
	}
}

func (t *TenantHandler) UnaryServerInterceptor(cfg *conf.Configuration) grpc.UnaryServerInterceptor {
	headerKey := getHeaderKey(cfg)
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		ctx, err = t.ExtractTenantID(ctx, headerKey)
		if err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

func (t *TenantHandler) StreamServerInterceptor(cfg *conf.Configuration) grpc.StreamServerInterceptor {
	headerKey := getHeaderKey(cfg)
	return func(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		newctx, err := t.ExtractTenantID(stream.Context(), headerKey)
		if err != nil {
			return err
		}
		ws := interceptor.WrapServerStream(stream)
		ws.WrappedContext = newctx
		return handler(srv, ws)
	}
}

type IdentityHandler struct{}

func (i *IdentityHandler) UnaryClientInterceptor(cfg *conf.Configuration) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if id, err := identity.UserIDFromContextAsInt(ctx); err == nil {
			ctx = metadata.AppendToOutgoingContext(ctx, identity.UserHeaderKey, strconv.Itoa(id))
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func (i *IdentityHandler) StreamClientInterceptor(cfg *conf.Configuration) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		if id, err := identity.UserIDFromContextAsInt(ctx); err == nil {
			ctx = metadata.AppendToOutgoingContext(ctx, identity.UserHeaderKey, strconv.Itoa(id))
		}
		return streamer(ctx, desc, cc, method, opts...)
	}
}

func (i *IdentityHandler) UnaryServerInterceptor(cfg *conf.Configuration) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if ids := md.Get(identity.UserHeaderKey); len(ids) > 0 {
				ctx = security.WithContext(ctx, security.NewGenericPrincipalByClaims(jwt.MapClaims{
					"sub": ids[0],
				}))
			}
		}
		return handler(ctx, req)
	}
}

func (i *IdentityHandler) StreamServerInterceptor(cfg *conf.Configuration) grpc.StreamServerInterceptor {
	return func(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		md, ok := metadata.FromIncomingContext(stream.Context())
		if !ok {
			return handler(srv, stream)
		}
		ids := md.Get(identity.UserHeaderKey)
		if len(ids) == 0 {
			return handler(srv, stream)
		}
		ctx := security.WithContext(stream.Context(), security.NewGenericPrincipalByClaims(jwt.MapClaims{
			"sub": ids[0],
		}))
		ws := interceptor.WrapServerStream(stream)
		ws.WrappedContext = ctx
		return handler(srv, ws)
	}
}
