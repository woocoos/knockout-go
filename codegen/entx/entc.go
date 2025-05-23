package entx

import (
	atlas "ariga.io/atlas/sql/schema"
	"embed"
	"entgo.io/contrib/entgql"
	"entgo.io/ent/dialect/sql/schema"
	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
	"github.com/vektah/gqlparser/v2/ast"
)

var (
	//go:embed template/*
	_templates embed.FS
)

// GlobalID is a global id template for Noder Query. Use with ChangeRelayNodeType().
//
// if you use GlobalID, you must use GID as a scalar type.
// and use ChangeRelayNodeType() in entgql.WithSchemaHook()
func GlobalID() entc.Option {
	return func(g *gen.Config) error {
		g.Templates = append(g.Templates, gen.MustParse(gen.NewTemplate("gql_globalid").
			Funcs(entgql.TemplateFuncs).
			ParseFS(_templates, "template/globalid.tmpl")))
		return nil
	}
}

func SimplePagination() entc.Option {
	return func(g *gen.Config) error {
		g.Templates = append(g.Templates, gen.MustParse(gen.NewTemplate("gql_pagination_simple").
			Funcs(entgql.TemplateFuncs).
			ParseFS(_templates, "template/gql_pagination_simple.tmpl")))
		return nil
	}
}

// ChangeRelayNodeType is a schema hook for a change relay node type to GID. Use with GlobalID().
//
// add it to entgql.WithSchemaHook()
func ChangeRelayNodeType() entgql.SchemaHook {
	idType := ast.NonNullNamedType("GID", nil)
	found := false
	return func(graph *gen.Graph, schema *ast.Schema) error {
		for _, field := range schema.Types["Query"].Fields {
			if field.Name == "node" {
				field.Arguments[0].Type = idType
				found = true
			}
			if field.Name == "nodes" {
				field.Arguments[0].Type = ast.NonNullListType(idType, nil)
				found = true
			}
		}
		if found && schema.Types["GID"] == nil {
			schema.Types["GID"] = &ast.Definition{
				Kind:        ast.Scalar,
				Name:        "GID",
				Description: "An object with a Global ID,for using in Noder interface.",
				Directives: ast.DirectiveList{&ast.Directive{
					Name:     "goModel",
					Location: ast.LocationObject,
					Arguments: ast.ArgumentList{
						{
							Name: "model",
							Value: &ast.Value{
								Kind: ast.StringValue,
								Raw:  "github.com/99designs/gqlgen/graphql.ID",
							},
						},
					},
				}},
			}
		}
		return nil
	}
}

// WithGqlWithTemplates is a schema hook for replace entgql default template.
// Note: this option must put before WithWhereInputs or which changed entgql templates option.
//
// extensions:
//  1. NodeTemplate:
//     Noder: add entcache context
func WithGqlWithTemplates() entgql.ExtensionOption {
	nodeTpl := gen.MustParse(gen.NewTemplate("node").
		Funcs(entgql.TemplateFuncs).ParseFS(_templates, "template/node.tmpl", "template/pagination.tmpl"))
	return entgql.WithTemplates(append(entgql.AllTemplates, nodeTpl)...)
}

// SkipTablesDiffHook is a schema migration hook for skip tables diff thus skip migration.
// the table name is database name,not the ent schema struct name.
//
//	err = client.Schema.Create(ctx,SkipTablesDiffHook("table1","table2"))
func SkipTablesDiffHook(tables ...string) schema.MigrateOption {
	return schema.WithDiffHook(func(next schema.Differ) schema.Differ {
		return schema.DiffFunc(func(current, desired *atlas.Schema) ([]atlas.Change, error) {
			var dt []*atlas.Table
		LOOP:
			for i, table := range desired.Tables {
				for _, t := range tables {
					if table.Name == t {
						continue LOOP
					}
				}
				dt = append(dt, desired.Tables[i])
			}
			desired.Tables = dt
			// Before calculating changes.
			changes, err := next.Diff(current, desired)
			if err != nil {
				return nil, err
			}
			// After diff, you can filter
			// changes or return new ones.
			return changes, nil
		})
	})
}
