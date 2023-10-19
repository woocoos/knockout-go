// Package snowflake provides a very simple Twitter snowflake generator.
//
// init初始化:
//
//	通过环境变量SNOWFLAKE_NODE_LIST和HOST_IP来确定节点ID.对于单机部署Node=1.
//	分布式部署时,节点通过主机IP匹配SNOWFLAKE_NODE_LIST中的索引位置(index起始为1),确定节点ID.如果未能匹配,则同单机部署ID=1.
//
// 同时支持使用配置文件初始化:
//
//	nodeBits: 节点位数,默认3,最多支持2^3 8个节点
//	stepBits: 序列号位数,默认8, 一毫秒内最多生成2^8 256个ID
//	nodeID: 节点ID,默认1
//
// 在使用默认值情况下,最多部署 8 台机器,每毫秒产生 8X 256 = 2048 个 ID 在一般的系统中已足够使用.
// 通过环境变量`SNOWFLAKE_DEFAULT`= true (1,t,T等) 来使用标准算法,
//
// K8S环境注意事项:
// 由于每个容器是独立的程序,因此Deployment 开启多实例时,不同的POD可能部署至同一台服务器时,会出现相同的节点ID,导致ID重复.尽量使用亲合性配置防止此情况.
// 针对不同的业务,一般操作的数据表也不同,ID 重复是可容忍的.
// 当机器增加时,只做增量配置变更,原运行的实例不需要重启.
package snowflake

import (
	"github.com/bwmarrin/snowflake"
	"github.com/tsingsun/woocoo/pkg/conf"
	"os"
	"strconv"
	"strings"
)

var (
	defaultNode *snowflake.Node
	// 时间戳起始时间 北京时间 2023-01-01 00:00:00
	epoch int64 = 1672531200000
)

const (
	nodeIDKey = "nodeID"
)

func init() {
	if err := SetDefaultNodeFromEnv(); err != nil {
		panic(err)
	}
}

func SetDefaultNodeFromEnv() (err error) {
	nodeID := 1
	if useDefault, _ := strconv.ParseBool(os.Getenv("SNOWFLAKE_DEFAULT")); !useDefault {
		snowflake.Epoch = epoch
		snowflake.NodeBits = 3
		snowflake.StepBits = 8
	}
	nodelist := strings.Split(os.Getenv("SNOWFLAKE_NODE_LIST"), ",")
	if len(nodelist) > 0 {
		hostip := os.Getenv("HOST_IP")
		for i, v := range nodelist {
			if v == hostip {
				nodeID = i + 1
				break
			}
		}
	}
	cnf := conf.NewFromStringMap(map[string]any{
		nodeIDKey: nodeID,
	})
	err = SetDefaultNode(cnf)
	return err
}

func SetDefaultNode(cnf *conf.Configuration) (err error) {
	if v := cnf.Int("nodeBits"); v > 0 {
		snowflake.NodeBits = uint8(v)
	}
	if v := cnf.Int("stepBits"); v > 0 {
		snowflake.StepBits = uint8(v)
	}
	if v := cnf.Int("epoch"); v > 0 {
		snowflake.Epoch = int64(v)
	}
	nid := cnf.Int(nodeIDKey)
	if nid <= 0 {
		nid = 1
	}
	defaultNode, err = snowflake.NewNode(int64(nid))
	return err
}

func New() snowflake.ID {
	return defaultNode.Generate()
}
