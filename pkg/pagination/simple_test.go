package pagination

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tsingsun/woocoo/pkg/gds"
)

func TestNeedLimit(t *testing.T) {
	type args struct {
		first  *int
		last   *int
		maxRow int
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Both parameters are nil",
			args: args{
				first:  nil,
				last:   nil,
				maxRow: 0,
			},
			wantErr: assert.Error,
		},
		{
			name: "First parameter exceeds max row",
			args: args{
				first:  gds.Ptr(101),
				last:   nil,
				maxRow: 100,
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorIs(t, err, ErrGreaterThanMaxRow)
			},
		},
		{
			name: "Last parameter exceeds max row",
			args: args{
				first:  nil,
				last:   gds.Ptr(101),
				maxRow: 100,
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorIs(t, err, ErrGreaterThanMaxRow)
			},
		},
		{
			name: "Both parameters exceed max row",
			args: args{
				first:  gds.Ptr(101),
				last:   gds.Ptr(102),
				maxRow: 100,
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorIs(t, err, ErrGreaterThanMaxRow)
			},
		},
		{
			name: "Both parameters within range",
			args: args{
				first:  gds.Ptr(50),
				last:   gds.Ptr(50),
				maxRow: 100,
			},
			wantErr: assert.NoError,
		},
		{
			name: "First parameter equals max row",
			args: args{
				first:  gds.Ptr(100),
				last:   nil,
				maxRow: 100,
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NeedLimit(tt.args.first, tt.args.last, tt.args.maxRow)
			tt.wantErr(t, got, "NeedLimit()")
		})
	}
}
