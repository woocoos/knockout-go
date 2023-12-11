package main

import (
	"github.com/99designs/gqlgen/api"
	"github.com/99designs/gqlgen/codegen/config"
	"github.com/woocoos/knockout-go/codegen/gqlx"
	"log"
	"os"
)

func main() {
	cfg, err := config.LoadConfig("./codegen/gqlgen/gqlgen.yml")
	if err != nil {
		log.Print("failed to load config", err.Error())
		os.Exit(2)
	}

	err = api.Generate(cfg,
		api.AddPlugin(gqlx.NewResolverPlugin(gqlx.WithRelayNodeEx(), gqlx.WithConfig(cfg))),
	)

	if err != nil {
		log.Print(err.Error())
		os.Exit(3)
	}
}
