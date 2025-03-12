package entx

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
	"entgo.io/ent/schema/field"
	"github.com/vektah/gqlparser/v2/ast"
	"text/template"
)

// DecimalExtension 修正声明为fieldx.Decimal时, 验证器因此外部类型,无法识别,需要进行修正.
//
// 如 对于Decimal类型,默认生成为:
//
//	if v, ok := oc.mutation.OrderQty(); ok {
//		if err := order.OrderQtyValidator(v.String()); err != nil {
//			return &ValidationError{Name: "order_qty", err: fmt.Errorf(`ent: validator failed for field "Order.order_qty": %w`, err)}
//		}
//	}
//
// 验证器函数应修改为类型参数为decimal.Decimal:
//
//	if v, ok := oc.mutation.OrderQty(); ok {
//		if err := decimal.OrderQtyValidator(v); err != nil {
//			return &ValidationError{Name: "order_qty", err: fmt.Errorf(`ent: validator failed for field "Order.order_qty": %w`, err)}
//		}
//	}
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

// DecimalHook 修正Decimal类型, 将类型修改为float64,以便将decimal.Decimal类型当成数值类型.
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

// DecimalScalar 注册decimal.Decimal类型为GraphQL Scalar类型.
func DecimalScalar() entgql.SchemaHook {
	return func(graph *gen.Graph, schema *ast.Schema) error {
		if schema.Types["Decimal"] == nil {
			schema.Types["Decimal"] = &ast.Definition{
				Kind:        ast.Scalar,
				Name:        "Decimal",
				Description: "Arbitrary-precision fixed-point decimal numbers",
			}
		}
		return nil
	}
}
