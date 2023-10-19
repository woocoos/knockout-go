package snowflake

import (
	"github.com/bwmarrin/snowflake"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tsingsun/woocoo/pkg/conf"
	"os"
	"testing"
)

func reset() {
	snowflake.Epoch = 1288834974657
	snowflake.NodeBits = 10
	snowflake.StepBits = 12
	defaultNode = nil
}

func TestSetDefaultNode(t *testing.T) {
	type args struct {
		cnf *conf.Configuration
	}
	tests := []struct {
		name  string
		args  args
		panic bool
		check func()
	}{
		{
			name: "init",
			args: args{
				cnf: nil,
			},
			check: func() {
				id := New()
				assert.EqualValues(t, 3, int(snowflake.NodeBits))
				assert.EqualValues(t, 8, int(snowflake.StepBits))
				assert.Len(t, id.String(), 14)
			},
		},
		{
			name: "default",
			args: args{
				cnf: conf.NewFromStringMap(map[string]any{}),
			},
			panic: false,
			check: func() {
				id := New()
				assert.Len(t, id.String(), 14)
			},
		},
		{
			name: "small",
			args: args{
				cnf: func() *conf.Configuration {
					return conf.NewFromStringMap(map[string]any{
						"nodeBits": 1,
						"stepBits": 8,
					})
				}(),
			},
			panic: false,
			check: func() {
				assert.Equal(t, uint8(1), snowflake.NodeBits)
				assert.Equal(t, uint8(8), snowflake.StepBits)
				id := New()
				assert.Len(t, id.String(), 14)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.panic {
				assert.Error(t, SetDefaultNode(tt.args.cnf))
				return
			}
			if tt.args.cnf != nil {
				require.NoError(t, SetDefaultNode(tt.args.cnf))
			}
			tt.check()
		})
	}
}

func TestSetDefaultNodeFromEnv(t *testing.T) {
	tests := []struct {
		name    string
		init    func()
		panic   bool
		check   func()
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "node from env",
			init: func() {
				require.NoError(t, os.Setenv("SNOWFLAKE_NODE_LIST", "127.0.0.1,10.0.0.1"))
				require.NoError(t, os.Setenv("HOST_IP", "10.0.0.1"))
				require.NoError(t, os.Setenv("SNOWFLAKE_DEFAULT", "1"))
			},
			panic: false,
			check: func() {
				assert.EqualValues(t, 10, snowflake.NodeBits)
				assert.EqualValues(t, 12, snowflake.StepBits)
				id := New()
				assert.EqualValues(t, 2, id.Node())
			},
		},
		{
			name: "env with default",
			init: func() {
				require.NoError(t, os.Setenv("SNOWFLAKE_NODE_LIST", "127.0.0.1,10.0.0.1"))
				require.NoError(t, os.Setenv("HOST_IP", "10.0.0.1"))
				require.NoError(t, os.Setenv("SNOWFLAKE_DEFAULT", "0"))
			},
			panic: false,
			check: func() {
				assert.EqualValues(t, 3, snowflake.NodeBits)
				assert.EqualValues(t, 8, snowflake.StepBits)
				id := New()
				assert.EqualValues(t, 2, id.Node())
			},
		},
		{
			name: "empty list",
			init: func() {
				require.NoError(t, os.Setenv("SNOWFLAKE_NODE_LIST", ""))
				require.NoError(t, os.Setenv("HOST_IP", "10.0.0.1"))
			},
			panic: false,
			check: func() {
				id := New()
				assert.EqualValues(t, 1, id.Node())
			},
		},
		{
			name: "empty list",
			init: func() {
				require.NoError(t, os.Setenv("SNOWFLAKE_NODE_LIST", "127.0.0.1,10.0.0.1"))
				require.NoError(t, os.Setenv("HOST_IP", ""))
			},
			panic: false,
			check: func() {
				id := New()
				assert.EqualValues(t, 1, id.Node())
			},
		},
	}
	for _, tt := range tests {
		reset()
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				os.Setenv("SNOWFLAKE_NODE_LIST", "")
				os.Setenv("HOST_IP", "")
				os.Setenv("SNOWFLAKE_DEFAULT", "")
			}()
			if tt.init != nil {
				tt.init()
			}
			if tt.panic {
				assert.Error(t, SetDefaultNodeFromEnv())
				return
			}
			require.NoError(t, SetDefaultNodeFromEnv())
			tt.check()
		})
	}
}
