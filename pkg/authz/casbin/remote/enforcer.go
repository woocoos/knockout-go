package remote

import (
	"context"

	"github.com/casbin/casbin/v2"
	"github.com/tsingsun/woocoo/pkg/conf"
	"github.com/woocoos/knockout-go/pkg/authz/casbin/proto"
)

// GrpcEnforcer 做为woocoo的验证实现支持,使用grpc调用远程服务获取底层数据.
// 注意,并没有实现所有接口, 需要特别注意使用方式.
type GrpcEnforcer struct {
	casbin.IEnforcer
	client *Client
}

func NewGrpcEnforcer(cfg *conf.Configuration) (ef *GrpcEnforcer, err error) {
	ef = &GrpcEnforcer{}
	ef.client, err = NewClient(cfg)
	return ef, err
}

// Enforce 实现casbin.IEnforcer接口
func (e *GrpcEnforcer) Enforce(params ...any) (bool, error) {
	var data []string
	for _, item := range params {
		data = append(data, item.(string))
	}
	res, err := e.client.Enforce(context.Background(), &proto.EnforceRequest{
		Params: data,
	})
	if err != nil {
		return false, err
	}
	return res.Res, err
}

func (e *GrpcEnforcer) GetImplicitPermissionsForUser(user string, domain ...string) ([][]string, error) {
	res, err := e.client.GetImplicitPermissionsForUser(context.Background(), &proto.PermissionRequest{
		User:   user,
		Domain: domain,
	})
	if err != nil {
		return nil, err
	}
	return replyTo2DSlice(res), nil
}

// replyTo2DSlice transforms a Array2DReply to a 2d string slice.
func replyTo2DSlice(reply *proto.Array2DReply) [][]string {
	result := make([][]string, 0)
	for _, value := range reply.D2 {
		result = append(result, value.D1)
	}
	return result
}
