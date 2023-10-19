package gql

import (
	"github.com/99designs/gqlgen/api"
	"github.com/99designs/gqlgen/codegen/config"
	"github.com/stretchr/testify/suite"
	"github.com/woocoos/knockout-go/test"
	"os"
	"path/filepath"
	"testing"
)

type gqlgenSuite struct {
	suite.Suite
	GqlConfig *config.Config
}

// set up
func (s *gqlgenSuite) SetupSuite() {
}

func TestGqlgen(t *testing.T) {
	suite.Run(t, new(gqlgenSuite))
}

func (s *gqlgenSuite) TestImplement() {
	gqlfile := filepath.Join(test.Path("../"), "./integration/gentest/testdata/resolver.yml")
	testdir := filepath.Dir(gqlfile)
	err := os.RemoveAll(filepath.Join(testdir, "tmp"))
	s.Require().NoError(os.Chdir(testdir))
	cfg, err := config.LoadConfig(gqlfile)
	s.Require().NoError(err)
	err = api.Generate(cfg, api.AddPlugin(NewResolverPlugin(WithRelayNodeEx())))
	s.Require().NoError(err)
}
