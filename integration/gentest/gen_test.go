package gentest

import (
	"github.com/99designs/gqlgen/api"
	"github.com/99designs/gqlgen/codegen/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/woocoos/knockout-go/codegen/gqlx"
	"os"
	"path/filepath"
	"testing"
)

func TestGen(t *testing.T) {
	gqlfile, err := filepath.Abs("./testdata/resolver.yml")
	require.NoError(t, err)
	testdir := filepath.Dir(gqlfile)
	err = os.RemoveAll(filepath.Join(testdir, "tmp"))
	assert.NoError(t, os.Chdir(testdir))
	cfg, err := config.LoadConfig(gqlfile)
	require.NoError(t, err)
	// the test generate new code field always, so no need adding config
	err = api.Generate(cfg, api.AddPlugin(gqlx.NewResolverPlugin(gqlx.WithRelayNodeEx(), gqlx.WithConfig(cfg))))
	if assert.NoError(t, err, "generate success then clean up") {
		if err = os.RemoveAll(filepath.Join(testdir, "tmp")); err != nil {
			t.Log(err)
		}
	}
}
