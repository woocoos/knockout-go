package remote

import (
	"github.com/tsingsun/woocoo/pkg/conf"
	"github.com/tsingsun/woocoo/rpc/grpcx"
	"github.com/woocoos/knockout-go/pkg/authz/casbin/proto"
)

// Client casbin客户端实现
type Client struct {
	proto.CasbinClient
}

func NewClient(cfg *conf.Configuration) (*Client, error) {
	cli, err := grpcx.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	conn, err := cli.Dial("")
	if err != nil {
		return nil, err
	}
	return &Client{
		CasbinClient: proto.NewCasbinClient(conn),
	}, nil
}
