// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/woocoos/knockout-go/integration/helloapp/ent/world"
)

// WorldCreate is the builder for creating a World entity.
type WorldCreate struct {
	config
	mutation *WorldMutation
	hooks    []Hook
}

// SetCreatedBy sets the "created_by" field.
func (wc *WorldCreate) SetCreatedBy(i int) *WorldCreate {
	wc.mutation.SetCreatedBy(i)
	return wc
}

// SetCreatedAt sets the "created_at" field.
func (wc *WorldCreate) SetCreatedAt(t time.Time) *WorldCreate {
	wc.mutation.SetCreatedAt(t)
	return wc
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (wc *WorldCreate) SetNillableCreatedAt(t *time.Time) *WorldCreate {
	if t != nil {
		wc.SetCreatedAt(*t)
	}
	return wc
}

// SetUpdatedBy sets the "updated_by" field.
func (wc *WorldCreate) SetUpdatedBy(i int) *WorldCreate {
	wc.mutation.SetUpdatedBy(i)
	return wc
}

// SetNillableUpdatedBy sets the "updated_by" field if the given value is not nil.
func (wc *WorldCreate) SetNillableUpdatedBy(i *int) *WorldCreate {
	if i != nil {
		wc.SetUpdatedBy(*i)
	}
	return wc
}

// SetUpdatedAt sets the "updated_at" field.
func (wc *WorldCreate) SetUpdatedAt(t time.Time) *WorldCreate {
	wc.mutation.SetUpdatedAt(t)
	return wc
}

// SetNillableUpdatedAt sets the "updated_at" field if the given value is not nil.
func (wc *WorldCreate) SetNillableUpdatedAt(t *time.Time) *WorldCreate {
	if t != nil {
		wc.SetUpdatedAt(*t)
	}
	return wc
}

// SetDeletedAt sets the "deleted_at" field.
func (wc *WorldCreate) SetDeletedAt(t time.Time) *WorldCreate {
	wc.mutation.SetDeletedAt(t)
	return wc
}

// SetNillableDeletedAt sets the "deleted_at" field if the given value is not nil.
func (wc *WorldCreate) SetNillableDeletedAt(t *time.Time) *WorldCreate {
	if t != nil {
		wc.SetDeletedAt(*t)
	}
	return wc
}

// SetTenantID sets the "tenant_id" field.
func (wc *WorldCreate) SetTenantID(i int) *WorldCreate {
	wc.mutation.SetTenantID(i)
	return wc
}

// SetName sets the "name" field.
func (wc *WorldCreate) SetName(s string) *WorldCreate {
	wc.mutation.SetName(s)
	return wc
}

// SetPowerBy sets the "power_by" field.
func (wc *WorldCreate) SetPowerBy(s string) *WorldCreate {
	wc.mutation.SetPowerBy(s)
	return wc
}

// SetNillablePowerBy sets the "power_by" field if the given value is not nil.
func (wc *WorldCreate) SetNillablePowerBy(s *string) *WorldCreate {
	if s != nil {
		wc.SetPowerBy(*s)
	}
	return wc
}

// SetID sets the "id" field.
func (wc *WorldCreate) SetID(i int) *WorldCreate {
	wc.mutation.SetID(i)
	return wc
}

// Mutation returns the WorldMutation object of the builder.
func (wc *WorldCreate) Mutation() *WorldMutation {
	return wc.mutation
}

// Save creates the World in the database.
func (wc *WorldCreate) Save(ctx context.Context) (*World, error) {
	if err := wc.defaults(); err != nil {
		return nil, err
	}
	return withHooks(ctx, wc.sqlSave, wc.mutation, wc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (wc *WorldCreate) SaveX(ctx context.Context) *World {
	v, err := wc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (wc *WorldCreate) Exec(ctx context.Context) error {
	_, err := wc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (wc *WorldCreate) ExecX(ctx context.Context) {
	if err := wc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (wc *WorldCreate) defaults() error {
	if _, ok := wc.mutation.CreatedAt(); !ok {
		if world.DefaultCreatedAt == nil {
			return fmt.Errorf("ent: uninitialized world.DefaultCreatedAt (forgotten import ent/runtime?)")
		}
		v := world.DefaultCreatedAt()
		wc.mutation.SetCreatedAt(v)
	}
	if _, ok := wc.mutation.PowerBy(); !ok {
		v := world.DefaultPowerBy
		wc.mutation.SetPowerBy(v)
	}
	return nil
}

// check runs all checks and user-defined validators on the builder.
func (wc *WorldCreate) check() error {
	if _, ok := wc.mutation.CreatedBy(); !ok {
		return &ValidationError{Name: "created_by", err: errors.New(`ent: missing required field "World.created_by"`)}
	}
	if _, ok := wc.mutation.CreatedAt(); !ok {
		return &ValidationError{Name: "created_at", err: errors.New(`ent: missing required field "World.created_at"`)}
	}
	if _, ok := wc.mutation.TenantID(); !ok {
		return &ValidationError{Name: "tenant_id", err: errors.New(`ent: missing required field "World.tenant_id"`)}
	}
	if _, ok := wc.mutation.Name(); !ok {
		return &ValidationError{Name: "name", err: errors.New(`ent: missing required field "World.name"`)}
	}
	return nil
}

func (wc *WorldCreate) sqlSave(ctx context.Context) (*World, error) {
	if err := wc.check(); err != nil {
		return nil, err
	}
	_node, _spec := wc.createSpec()
	if err := sqlgraph.CreateNode(ctx, wc.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	if _spec.ID.Value != _node.ID {
		id := _spec.ID.Value.(int64)
		_node.ID = int(id)
	}
	wc.mutation.id = &_node.ID
	wc.mutation.done = true
	return _node, nil
}

func (wc *WorldCreate) createSpec() (*World, *sqlgraph.CreateSpec) {
	var (
		_node = &World{config: wc.config}
		_spec = sqlgraph.NewCreateSpec(world.Table, sqlgraph.NewFieldSpec(world.FieldID, field.TypeInt))
	)
	if id, ok := wc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = id
	}
	if value, ok := wc.mutation.CreatedBy(); ok {
		_spec.SetField(world.FieldCreatedBy, field.TypeInt, value)
		_node.CreatedBy = value
	}
	if value, ok := wc.mutation.CreatedAt(); ok {
		_spec.SetField(world.FieldCreatedAt, field.TypeTime, value)
		_node.CreatedAt = value
	}
	if value, ok := wc.mutation.UpdatedBy(); ok {
		_spec.SetField(world.FieldUpdatedBy, field.TypeInt, value)
		_node.UpdatedBy = value
	}
	if value, ok := wc.mutation.UpdatedAt(); ok {
		_spec.SetField(world.FieldUpdatedAt, field.TypeTime, value)
		_node.UpdatedAt = value
	}
	if value, ok := wc.mutation.DeletedAt(); ok {
		_spec.SetField(world.FieldDeletedAt, field.TypeTime, value)
		_node.DeletedAt = value
	}
	if value, ok := wc.mutation.TenantID(); ok {
		_spec.SetField(world.FieldTenantID, field.TypeInt, value)
		_node.TenantID = value
	}
	if value, ok := wc.mutation.Name(); ok {
		_spec.SetField(world.FieldName, field.TypeString, value)
		_node.Name = value
	}
	if value, ok := wc.mutation.PowerBy(); ok {
		_spec.SetField(world.FieldPowerBy, field.TypeString, value)
		_node.PowerBy = value
	}
	return _node, _spec
}

// WorldCreateBulk is the builder for creating many World entities in bulk.
type WorldCreateBulk struct {
	config
	err      error
	builders []*WorldCreate
}

// Save creates the World entities in the database.
func (wcb *WorldCreateBulk) Save(ctx context.Context) ([]*World, error) {
	if wcb.err != nil {
		return nil, wcb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(wcb.builders))
	nodes := make([]*World, len(wcb.builders))
	mutators := make([]Mutator, len(wcb.builders))
	for i := range wcb.builders {
		func(i int, root context.Context) {
			builder := wcb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*WorldMutation)
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
					_, err = mutators[i+1].Mutate(root, wcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, wcb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, wcb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (wcb *WorldCreateBulk) SaveX(ctx context.Context) []*World {
	v, err := wcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (wcb *WorldCreateBulk) Exec(ctx context.Context) error {
	_, err := wcb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (wcb *WorldCreateBulk) ExecX(ctx context.Context) {
	if err := wcb.Exec(ctx); err != nil {
		panic(err)
	}
}
