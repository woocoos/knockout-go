package entx

import (
	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGlobalID(t *testing.T) {
	tests := []struct {
		name  string
		check entc.Option
	}{
		{
			name: "test",
			check: func(g *gen.Config) error {
				for _, template := range g.Templates {
					if template.Name() == "gql_globalid" {
						return nil
					}
				}
				assert.FailNow(t, "template not found")
				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GlobalID()
			cfg := &gen.Config{}
			assert.NoError(t, got(cfg))
		})
	}
}
