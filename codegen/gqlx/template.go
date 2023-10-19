package gql

import (
	"embed"
	"github.com/99designs/gqlgen/codegen"
	"github.com/vektah/gqlparser/v2/ast"
	"go/types"
	"strings"
	"text/template"
)

var (
	//go:embed template
	templateDir embed.FS
	funcs       = template.FuncMap{
		"hasPrefix":        strings.HasPrefix,
		"isEntCreate":      isEntCreate,
		"isEntUpdate":      isEntUpdate,
		"isEntDelete":      isEntDelete,
		"findCreateEntArg": findCreateEntArgs,
		"findUpdateEntArg": findUpdateEntArgs,
		"toLower":          strings.ToLower, // some as ent
		"trimSuffix":       strings.TrimSuffix,
		"trimPrefix":       strings.TrimPrefix,
	}
)

type MutationType int

const (
	MutationCreate MutationType = iota
	MutationUpdate
	MutationDelete
)

func isEntCreate(field *codegen.Field) bool {
	if field.Object.Definition.Name != "Mutation" {
		return false
	}
	if len(field.Args) != 1 {
		return false
	}
	if field.Args[0].TypeReference.Definition.Kind != ast.InputObject {
		return false
	}
	switch {
	case strings.HasPrefix(field.GoFieldName, "Create"):
		return true
	}
	return false
}

func isEntUpdate(field *codegen.Field) bool {
	if field.Object.Definition.Name != "Mutation" {
		return false
	}
	if len(field.Args) != 2 {
		return false
	}
	switch {
	case strings.HasPrefix(field.GoFieldName, "Update"):
		return true
	}
	return false
}

func isEntDelete(field *codegen.Field) bool {
	if field.Object.Definition.Name != "Mutation" {
		return false
	}

	switch {
	case strings.HasPrefix(field.GoFieldName, "Delete"):
		// cannot bool pointer
		if field.TypeReference.GO.String() != types.Typ[types.Bool].String() {
			return false
		}
		return true
	}
	return false
}

func findCreateEntArgs(field *codegen.Field) string {
	return findEntArg(field, MutationCreate)
}

func findUpdateEntArgs(field *codegen.Field) (input string) {
	return findEntArg(field, MutationUpdate)
}

func findEntArg(field *codegen.Field, mt MutationType) (input string) {
	entTypeName := field.TypeReference.Definition.Name
	for _, arg := range field.Args {
		switch arg.TypeReference.Definition.Kind {
		case ast.InputObject:
			switch mt {
			case MutationCreate, MutationUpdate, MutationDelete:
				if strings.HasSuffix(arg.Type.NamedType, entTypeName+"Input") {
					input = arg.Name
					return
				}
			}
		}
	}
	return
}

func isPagination(field *codegen.Field) bool {
	if len(field.Args) != 6 {
		return false
	}
	count := 0
	for _, arg := range field.Args {
		switch arg.Name {
		case "first", "after", "last", "before", "where", "orderBy":
			count++
		}
	}
	return count == 6
}
