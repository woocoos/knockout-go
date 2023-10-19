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
	err = api.Generate(cfg, api.AddPlugin(gqlx.NewResolverPlugin(gqlx.WithRelayNodeEx())))
	assert.NoError(t, err)
}