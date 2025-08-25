package idgrpc

import (
	"context"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
	"github.com/tsingsun/woocoo/pkg/conf"
	"github.com/tsingsun/woocoo/pkg/security"
	"github.com/tsingsun/woocoo/rpc/grpcx"
	"github.com/tsingsun/woocoo/rpc/grpcx/interceptor"
	"github.com/woocoos/knockout-go/pkg/identity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	tenantName    = "tenant"
	domainName    = "domain"
	headerKeyPath = "headerKey"
	userName      = "user"
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
	dh := &DomainHandler{}
	grpcx.RegisterUnaryClientInterceptor(domainName, dh.UnaryClientInterceptor)
	grpcx.RegisterStreamClientInterceptor(domainName, dh.StreamClientInterceptor)
	grpcx.RegisterGrpcUnaryInterceptor(domainName, dh.UnaryServerInterceptor)
	grpcx.RegisterGrpcStreamInterceptor(domainName, dh.StreamServerInterceptor)
}

// TenantHandler is a grpc interceptor for tenant id.
type TenantHandler struct{}

// extractTenantID extracts the tenant ID from the metadata and returns the updated context.
// bool indicates whether the tenant ID is found in the metadata.
func (h TenantHandler) extractTenantID(ctx context.Context, headerKey string) (context.Context, bool, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx, false, nil
	}
	ids := md.Get(headerKey)
	if len(ids) == 0 {
		return ctx, false, nil
	}
	id, err := strconv.Atoi(ids[0])
	if err != nil {
		return ctx, true, err
	}
	return identity.WithTenantID(ctx, id), true, nil
}

func (h TenantHandler) getHeaderKey(cfg *conf.Configuration) string {
	headerKey := cfg.String(headerKeyPath)
	if headerKey == "" {
		headerKey = identity.TenantHeaderKey
	}
	return headerKey
}

func (h TenantHandler) UnaryClientInterceptor(cfg *conf.Configuration) grpc.UnaryClientInterceptor {
	headerKey := h.getHeaderKey(cfg)
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if id, ok := identity.TenantIDLoadFromContext(ctx); ok {
			ctx = metadata.AppendToOutgoingContext(ctx, headerKey, strconv.Itoa(id))
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func (h TenantHandler) StreamClientInterceptor(cfg *conf.Configuration) grpc.StreamClientInterceptor {
	headerKey := h.getHeaderKey(cfg)
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		if id, ok := identity.TenantIDLoadFromContext(ctx); ok {
			ctx = metadata.AppendToOutgoingContext(ctx, headerKey, strconv.Itoa(id))
		}
		return streamer(ctx, desc, cc, method, opts...)
	}
}

func (h TenantHandler) UnaryServerInterceptor(cfg *conf.Configuration) grpc.UnaryServerInterceptor {
	headerKey := h.getHeaderKey(cfg)
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		ctx, _, err = h.extractTenantID(ctx, headerKey)
		if err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

func (h TenantHandler) StreamServerInterceptor(cfg *conf.Configuration) grpc.StreamServerInterceptor {
	headerKey := h.getHeaderKey(cfg)
	return func(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx, exist, err := h.extractTenantID(stream.Context(), headerKey)
		if err != nil {
			return err
		} else if exist {
			ws := interceptor.WrapServerStream(stream)
			ws.WrappedContext = ctx
			return handler(srv, ws)
		}
		return handler(srv, stream)
	}
}

// IdentityHandler is a grpc interceptor for identity. Be careful, server interceptor call will conflict with grpc JWT interceptor
// because they call the same context setter method `security.WithContext`, but this only pass identity id.
type IdentityHandler struct{}

func (IdentityHandler) getHeaderKey(cfg *conf.Configuration) string {
	headerKey := cfg.String(headerKeyPath)
	if headerKey == "" {
		headerKey = identity.UserHeaderKey
	}
	return headerKey
}

// extractIdentityID extracts the identity ID from the metadata and returns the updated context.
func (h IdentityHandler) extractIdentityID(ctx context.Context, headerKey string) (context.Context, bool, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx, false, nil
	}
	ids := md.Get(headerKey)
	if len(ids) == 0 {
		return ctx, false, nil
	}
	ctx = security.WithContext(ctx, security.NewGenericPrincipalByClaims(jwt.MapClaims{
		"sub": ids[0],
	}))
	return ctx, true, nil
}

func (h IdentityHandler) UnaryClientInterceptor(cfg *conf.Configuration) grpc.UnaryClientInterceptor {
	headerKey := h.getHeaderKey(cfg)
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if id, err := identity.UserIDFromContextAsInt(ctx); err == nil {
			ctx = metadata.AppendToOutgoingContext(ctx, headerKey, strconv.Itoa(id))
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func (h IdentityHandler) StreamClientInterceptor(cfg *conf.Configuration) grpc.StreamClientInterceptor {
	headerKey := h.getHeaderKey(cfg)
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		if id, err := identity.UserIDFromContextAsInt(ctx); err == nil {
			ctx = metadata.AppendToOutgoingContext(ctx, headerKey, strconv.Itoa(id))
		}
		return streamer(ctx, desc, cc, method, opts...)
	}
}

func (h IdentityHandler) UnaryServerInterceptor(cfg *conf.Configuration) grpc.UnaryServerInterceptor {
	headerKey := h.getHeaderKey(cfg)
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		ctx, _, err = h.extractIdentityID(ctx, headerKey)
		if err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

func (h IdentityHandler) StreamServerInterceptor(cfg *conf.Configuration) grpc.StreamServerInterceptor {
	headerKey := h.getHeaderKey(cfg)
	return func(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx, exist, err := h.extractIdentityID(stream.Context(), headerKey)
		if err != nil {
			return err
		} else if exist {
			ws := interceptor.WrapServerStream(stream)
			ws.WrappedContext = ctx
			return handler(srv, ws)
		}
		return handler(srv, stream)
	}
}

// DomainHandler is a grpc interceptor for domain id.
type DomainHandler struct{}

func (DomainHandler) getHeaderKey(cfg *conf.Configuration) string {
	headerKey := cfg.String(headerKeyPath)
	if headerKey == "" {
		headerKey = identity.DomainHeaderKey
	}
	return headerKey
}

func (h DomainHandler) extractDomainID(ctx context.Context, headerKey string) (context.Context, bool, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx, false, nil
	}
	ids := md.Get(headerKey)
	if len(ids) == 0 {
		return ctx, false, nil
	}
	id, err := strconv.Atoi(ids[0])
	if err != nil {
		return ctx, true, err
	}
	return identity.WithDomainID(ctx, id), true, nil
}

func (h DomainHandler) UnaryClientInterceptor(cfg *conf.Configuration) grpc.UnaryClientInterceptor {
	headerKey := h.getHeaderKey(cfg)
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if id, ok := identity.DomainIDLoadFromContext(ctx); ok {
			ctx = metadata.AppendToOutgoingContext(ctx, headerKey, strconv.Itoa(id))
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func (h DomainHandler) StreamClientInterceptor(cfg *conf.Configuration) grpc.StreamClientInterceptor {
	headerKey := h.getHeaderKey(cfg)
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		if id, ok := identity.DomainIDLoadFromContext(ctx); ok {
			ctx = metadata.AppendToOutgoingContext(ctx, headerKey, strconv.Itoa(id))
		}
		return streamer(ctx, desc, cc, method, opts...)
	}
}

func (h DomainHandler) UnaryServerInterceptor(cfg *conf.Configuration) grpc.UnaryServerInterceptor {
	headerKey := h.getHeaderKey(cfg)
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		ctx, _, err = h.extractDomainID(ctx, headerKey)
		if err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

func (h DomainHandler) StreamServerInterceptor(cfg *conf.Configuration) grpc.StreamServerInterceptor {
	headerKey := h.getHeaderKey(cfg)
	return func(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx, exist, err := h.extractDomainID(stream.Context(), headerKey)
		if err != nil {
			return err
		} else if exist {
			ws := interceptor.WrapServerStream(stream)
			ws.WrappedContext = ctx
			return handler(srv, ws)
		}
		return handler(srv, stream)
	}
}
