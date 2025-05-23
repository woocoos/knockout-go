package gqlx

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/99designs/gqlgen/codegen"
	"github.com/99designs/gqlgen/codegen/config"
	"github.com/99designs/gqlgen/codegen/templates"
	"github.com/99designs/gqlgen/plugin"
	"golang.org/x/tools/imports"
	"io/fs"
	"os"
	"strings"
	"text/template"
	"time"
)

type Option func(*ResolverPlugin)

// WithRelayNodeEx enable relay node extended by globalID.
func WithRelayNodeEx() Option {
	return func(plugin *ResolverPlugin) {
		plugin.useRelayNodeEx = true
	}
}

// WithConfig set config for resolver plugin. in v0.17.41 need the config. so add this method for compatible.
func WithConfig(cfg *config.Config) Option {
	return func(plugin *ResolverPlugin) {
		plugin.config = cfg
	}
}

var (
	_ plugin.ResolverImplementer = (*ResolverPlugin)(nil)
	_ plugin.CodeGenerator       = (*ResolverPlugin)(nil)
)

type ResolverPlugin struct {
	config         *config.Config
	rewriter       *Rewriter
	resolverTpl    *template.Template
	useRelayNodeEx bool
}

func NewResolverPlugin(opt ...Option) *ResolverPlugin {
	r := &ResolverPlugin{
		resolverTpl: template.Must(template.New("resolver").
			Funcs(funcs).
			ParseFS(templateDir, "template/*.tmpl")),
	}
	for _, option := range opt {
		option(r)
	}
	return r
}

func (r *ResolverPlugin) Name() string {
	return "ent-resolver"
}

// Implement gqlgen api.ResolverImplementer
func (r *ResolverPlugin) Implement(prevImplementation string, f *codegen.Field) (val string) {
	var (
		err error
	)
	if r.rewriter == nil {
		r.rewriter, err = NewRewriter(r.config.Resolver.Dir())
		if err != nil {
			panic(err)
		}
	}
	if r.config != nil {
		rs := f.Object.ResolverInterface.String()
		sn := templates.LcFirst(rs[strings.LastIndex(rs, ".")+1:])
		implementation := strings.TrimSpace(r.rewriter.GetMethodBody(sn, f.GoFieldName))
		if implementation != "" {
			return implementation
		}
	}
	switch {
	case f.Object.Definition.Name == "Mutation":
		val, err = r.Mutation(f)
	case f.Object.Definition.Name == "Query":
		val, err = r.Query(f)
	default:
		return fmt.Sprintf("panic(fmt.Errorf(\"not implemented: %v - %v\"))", f.GoFieldName, f.Name)
	}
	if err != nil {
		panic(err)
	}
	return
}

// GenerateCode implement api.CodeGenerator
func (r *ResolverPlugin) GenerateCode(data *codegen.Data) error {
	fi, err := os.Stat(data.Config.Resolver.Filename)
	// just override the resolver.go in this time if file is new created.
	if errors.Is(err, fs.ErrNotExist) || time.Now().Sub(fi.ModTime()) < time.Second*5 {
		err := r.OverrideResolverStruct(data.Config)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *ResolverPlugin) FormatFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("format file:read file %s: %w", path, err)
	}
	src, err := imports.Process(path, content, nil)
	if err != nil {
		return fmt.Errorf("format file %s: %w", path, err)
	}
	if err := os.WriteFile(path, src, 0644); err != nil {
		return fmt.Errorf("format file:write file %s: %w", path, err)
	}
	return nil
}

func (r *ResolverPlugin) OverrideResolverStruct(config *config.Config) error {
	b := &bytes.Buffer{}
	err := r.resolverTpl.ExecuteTemplate(b, "resolver", config)
	if err != nil {
		return err
	}
	path := config.Resolver.Filename
	if path == "" { // no resolver file
		return nil
	}
	err = os.WriteFile(path, b.Bytes(), 0644)
	if err != nil {
		return fmt.Errorf("write file %s: %w", path, err)
	}
	if err = r.FormatFile(path); err != nil {
		return err
	}

	return nil
}

func (r *ResolverPlugin) Mutation(f *codegen.Field) (string, error) {
	var (
		b   = &bytes.Buffer{}
		err error
	)
	if f.Object.Definition.Name == "Mutation" {
		switch {
		case isEntCreate(f):
			err = r.resolverTpl.ExecuteTemplate(b, "ent-create", f)
		case isEntUpdate(f):
			err = r.resolverTpl.ExecuteTemplate(b, "ent-update", f)
		case isEntDelete(f):
			err = r.resolverTpl.ExecuteTemplate(b, "ent-delete", f)
		default:
			return fmt.Sprintf("panic(fmt.Errorf(\"not implemented: %v - %v\"))", f.GoFieldName, f.Name), nil
		}
	}
	return b.String(), err
}

func (r *ResolverPlugin) Query(f *codegen.Field) (string, error) {
	var (
		err error
		b   = &bytes.Buffer{}
	)
	switch f.FieldDefinition.Name {
	case "node":
		if r.useRelayNodeEx {
			err = r.resolverTpl.ExecuteTemplate(b, "node-ex", f)
		} else {
			err = r.resolverTpl.ExecuteTemplate(b, "node", f)
		}
	case "nodes":
		if r.useRelayNodeEx {
			err = r.resolverTpl.ExecuteTemplate(b, "nodes-ex", f)
		} else {
			err = r.resolverTpl.ExecuteTemplate(b, "nodes", f)
		}
	default:
		if isPagination(f) {
			err = r.resolverTpl.ExecuteTemplate(b, "pagination", f)
		} else {
			return fmt.Sprintf("panic(fmt.Errorf(\"not implemented: %v - %v\"))", f.GoFieldName, f.Name), nil
		}
	}
	return b.String(), err
}
