package entx

import (
	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
	"entgo.io/ent/schema/field"
	"text/template"
)

type DecimalExtension struct {
	entc.DefaultExtension
}

func (DecimalExtension) Hooks() []gen.Hook {
	return []gen.Hook{
		DecimalHook(),
	}
}

func (DecimalExtension) Func() template.FuncMap {
	return template.FuncMap{
		"isCustomerField": func(f *gen.Field) bool {
			if !f.HasGoType() {
				return false
			}
			if f.Type.Numeric() && f.Type.RType != nil && f.Type.RType.PkgPath == "github.com/shopspring/decimal" {
				return true
			}
			return false
		},
	}
}

func (d DecimalExtension) Templates() []*gen.Template {
	return []*gen.Template{
		gen.MustParse(gen.NewTemplate("runtime").
			Funcs(d.Func()).
			ParseFS(_templates, "template/runtime.tmpl")),
		gen.MustParse(gen.NewTemplate("meta").
			Funcs(d.Func()).
			ParseFS(_templates, "template/meta.tmpl")),
		gen.MustParse(gen.NewTemplate("create").
			Funcs(d.Func()).
			ParseFS(_templates, "template/create.tmpl")),
		gen.MustParse(gen.NewTemplate("update").
			Funcs(d.Func()).
			ParseFS(_templates, "template/update.tmpl")),
	}
}

func DecimalHook() gen.Hook {
	return func(next gen.Generator) gen.Generator {
		return gen.GenerateFunc(func(g *gen.Graph) error {
			for _, nodes := range g.Nodes {
				for _, f := range nodes.Fields {
					if f.Type.RType != nil && f.Type.RType.String() == "decimal.Decimal" {
						f.Type.Type = field.TypeFloat64
					}
				}
			}
			return next.Generate(g)
		})
	}
}
