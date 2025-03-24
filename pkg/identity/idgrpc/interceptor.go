package idgrpc

import (
	"context"
	"github.com/tsingsun/woocoo/pkg/conf"
	"github.com/tsingsun/woocoo/rpc/grpcx"
	"github.com/tsingsun/woocoo/rpc/grpcx/interceptor"
	"github.com/woocoos/knockout-go/pkg/identity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"strconv"
)

const (
	name          = "tenant"
	headerKeyPath = "headerKey"
)

func init() {
	th := &TenantHandler{}
	grpcx.RegisterUnaryClientInterceptor(name, th.UnaryClientInterceptor)
	grpcx.RegisterStreamClientInterceptor(name, th.StreamClientInterceptor)
	grpcx.RegisterGrpcUnaryInterceptor(name, th.UnaryServerInterceptor)
	grpcx.RegisterGrpcStreamInterceptor(name, th.StreamServerInterceptor)
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
	headerKey := cfg.String(headerKeyPath)
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
