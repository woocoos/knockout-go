// This package is used to Debug
package main

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
	"github.com/woocoos/knockout-go/codegen/entx"
	"log"
)

func main() {
	ex, err := entgql.NewExtension(
		entgql.WithSchemaGenerator(),
		entx.WithGqlWithTemplates(),
		entgql.WithWhereInputs(true),
		entgql.WithSchemaHook(entx.ChangeRelayNodeType()),
	)
	if err != nil {
		log.Fatalf("creating entgql extension: %v", err)
	}
	opts := []entc.Option{
		entc.Extensions(ex, entx.DecimalExtension{}),
		entx.GlobalID(),
		entx.SimplePagination(),
	}
	err = entc.Generate("./ent/schema", &gen.Config{
		Package: "github.com/woocoos/knockout-go/integration/nocache/ent",
		Target:  "./ent",
	},
		opts...)
	if err != nil {
		log.Fatalf("running ent codegen: %v", err)
	}
}
