//go:build ignore

package main

import (
	"log"

	"entgo.io/contrib/entgql"
	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
	entcachegen "github.com/woocoos/entcache/gen"
	"github.com/woocoos/knockout-go/codegen/entx"
)

func main() {
	ex, err := entgql.NewExtension(
		entgql.WithSchemaGenerator(),
		entx.WithGqlWithTemplates(),
		entgql.WithWhereInputs(true),
		entgql.WithConfigPath("./gqlgen.yml"),
		entgql.WithSchemaPath("./ent.graphql"),
		entgql.WithSchemaHook(entx.ChangeRelayNodeType(), entx.DecimalScalar()),
	)
	if err != nil {
		log.Fatalf("creating entgql extension: %v", err)
	}
	opts := []entc.Option{
		entc.Extensions(ex, entx.DecimalExtension{}),
		entx.GlobalID(),
		entx.SimplePagination(),
		entcachegen.QueryCache(),
	}
	err = entc.Generate("./ent/schema", &gen.Config{},
		opts...)
	if err != nil {
		log.Fatalf("running ent codegen: %v", err)
	}
}
