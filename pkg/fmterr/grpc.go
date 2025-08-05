package fmterr

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/tsingsun/woocoo/pkg/conf"
	"github.com/tsingsun/woocoo/rpc/grpcx"
	"github.com/tsingsun/woocoo/web/handler"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func init() {
	grpcx.RegisterGrpcUnaryInterceptor("errorHandler", UnaryServerInterceptor)
	grpcx.RegisterUnaryClientInterceptor("errorHandler", UnaryClientInterceptorInGin)
}

type statusError struct {
	Error string `json:"error"`
	Meta  any    `json:"meta"`
}

// WrapperGrpcStatus wrap grpc error to pass to client.
func WrapperGrpcStatus(err error) error {
	type grpcstatus interface{ GRPCStatus() *status.Status }
	if gs, ok := err.(grpcstatus); ok {
		code, txt := handler.LookupErrorCode(uint64(gs.GRPCStatus().Code()), err)
		if code > 0 {
			return status.Error(codes.Code(code), txt)
		}
	}
	type fmterr interface {
		self() *Error
	}
	if gerr, ok := err.(fmterr); ok {
		e := gerr.self()
		code, txt := handler.LookupErrorCode(uint64(e.Type), err)
		if code > 0 {
			if e.Meta != nil {
				e.Err = errors.New(txt)
				jt, je := json.Marshal(e.JSON())
				if je != nil {
					return status.Errorf(codes.Code(code), txt+" %w", je)
				}
				return status.Error(codes.Code(code), string(jt))
			}
			return status.Error(codes.Code(code), txt)
		}
	}
	return err
}

// UnaryServerInterceptor error handler for grpc server side
func UnaryServerInterceptor(cfg *conf.Configuration) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		resp, err = handler(ctx, req)
		if err != nil {
			werr := WrapperGrpcStatus(err)
			return resp, werr
		}
		return resp, err
	}
}

// UnaryClientInterceptorInGin error handler for grpc client side in Gin.
// It will try to convert grpc status to gin.Error.
func UnaryClientInterceptorInGin(_ *conf.Configuration) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		err := invoker(ctx, method, req, reply, cc, opts...)
		if err != nil {
			se, ok := status.FromError(err)
			msg := se.Message()
			if ok && messageIsJSONLoose(msg) {
				var ge statusError
				if je := json.Unmarshal([]byte(msg), &ge); je != nil {
					return err
				}
				return &gin.Error{
					Type: gin.ErrorType(se.Code()),
					Err:  errors.New(ge.Error),
					Meta: ge.Meta,
				}
			} else if ok {
				return &gin.Error{
					Type: gin.ErrorType(se.Code()),
					Err:  errors.New(msg),
				}
			}
		}
		return err
	}
}

// 判断是否json字符串,使用宽松的方式进行判断
func messageIsJSONLoose(str string) bool {
	// 检查是否为空
	if str == "" {
		return false
	}

	// 检查首尾字符是否为有效的JSON开始和结束符
	firstChar := str[0]
	lastChar := str[len(str)-1]

	if !(firstChar == '{' && lastChar == '}') {
		return false
	}
	return true
}
