// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"

	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/woocoos/knockout-go/integration/helloapp/ent/hello"
)

// HelloCreate is the builder for creating a Hello entity.
type HelloCreate struct {
	config
	mutation *HelloMutation
	hooks    []Hook
}

// SetTenantID sets the "tenant_id" field.
func (hc *HelloCreate) SetTenantID(i int) *HelloCreate {
	hc.mutation.SetTenantID(i)
	return hc
}

// SetName sets the "name" field.
func (hc *HelloCreate) SetName(s string) *HelloCreate {
	hc.mutation.SetName(s)
	return hc
}

// SetID sets the "id" field.
func (hc *HelloCreate) SetID(i int) *HelloCreate {
	hc.mutation.SetID(i)
	return hc
}

// Mutation returns the HelloMutation object of the builder.
func (hc *HelloCreate) Mutation() *HelloMutation {
	return hc.mutation
}

// Save creates the Hello in the database.
func (hc *HelloCreate) Save(ctx context.Context) (*Hello, error) {
	return withHooks(ctx, hc.sqlSave, hc.mutation, hc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (hc *HelloCreate) SaveX(ctx context.Context) *Hello {
	v, err := hc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (hc *HelloCreate) Exec(ctx context.Context) error {
	_, err := hc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (hc *HelloCreate) ExecX(ctx context.Context) {
	if err := hc.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (hc *HelloCreate) check() error {
	if _, ok := hc.mutation.TenantID(); !ok {
		return &ValidationError{Name: "tenant_id", err: errors.New(`ent: missing required field "Hello.tenant_id"`)}
	}
	if _, ok := hc.mutation.Name(); !ok {
		return &ValidationError{Name: "name", err: errors.New(`ent: missing required field "Hello.name"`)}
	}
	return nil
}

func (hc *HelloCreate) sqlSave(ctx context.Context) (*Hello, error) {
	if err := hc.check(); err != nil {
		return nil, err
	}
	_node, _spec := hc.createSpec()
	if err := sqlgraph.CreateNode(ctx, hc.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	if _spec.ID.Value != _node.ID {
		id := _spec.ID.Value.(int64)
		_node.ID = int(id)
	}
	hc.mutation.id = &_node.ID
	hc.mutation.done = true
	return _node, nil
}

func (hc *HelloCreate) createSpec() (*Hello, *sqlgraph.CreateSpec) {
	var (
		_node = &Hello{config: hc.config}
		_spec = sqlgraph.NewCreateSpec(hello.Table, sqlgraph.NewFieldSpec(hello.FieldID, field.TypeInt))
	)
	if id, ok := hc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = id
	}
	if value, ok := hc.mutation.TenantID(); ok {
		_spec.SetField(hello.FieldTenantID, field.TypeInt, value)
		_node.TenantID = value
	}
	if value, ok := hc.mutation.Name(); ok {
		_spec.SetField(hello.FieldName, field.TypeString, value)
		_node.Name = value
	}
	return _node, _spec
}

// HelloCreateBulk is the builder for creating many Hello entities in bulk.
type HelloCreateBulk struct {
	config
	err      error
	builders []*HelloCreate
}

// Save creates the Hello entities in the database.
func (hcb *HelloCreateBulk) Save(ctx context.Context) ([]*Hello, error) {
	if hcb.err != nil {
		return nil, hcb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(hcb.builders))
	nodes := make([]*Hello, len(hcb.builders))
	mutators := make([]Mutator, len(hcb.builders))
	for i := range hcb.builders {
		func(i int, root context.Context) {
			builder := hcb.builders[i]
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*HelloMutation)
				if !ok {
					return nil, fmt.Errorf("unexpected mutation type %T", m)
				}
				if err := builder.check(); err != nil {
					return nil, err
				}
				builder.mutation = mutation
				var err error
				nodes[i], specs[i] = builder.createSpec()
				if i < len(mutators)-1 {
					_, err = mutators[i+1].Mutate(root, hcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, hcb.driver, spec); err != nil {
						if sqlgraph.IsConstraintError(err) {
							err = &ConstraintError{msg: err.Error(), wrap: err}
						}
					}
				}
				if err != nil {
					return nil, err
				}
				mutation.id = &nodes[i].ID
				if specs[i].ID.Value != nil && nodes[i].ID == 0 {
					id := specs[i].ID.Value.(int64)
					nodes[i].ID = int(id)
				}
				mutation.done = true
				return nodes[i], nil
			})
			for i := len(builder.hooks) - 1; i >= 0; i-- {
				mut = builder.hooks[i](mut)
			}
			mutators[i] = mut
		}(i, ctx)
	}
	if len(mutators) > 0 {
		if _, err := mutators[0].Mutate(ctx, hcb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (hcb *HelloCreateBulk) SaveX(ctx context.Context) []*Hello {
	v, err := hcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (hcb *HelloCreateBulk) Exec(ctx context.Context) error {
	_, err := hcb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (hcb *HelloCreateBulk) ExecX(ctx context.Context) {
	if err := hcb.Exec(ctx); err != nil {
		panic(err)
	}
}
