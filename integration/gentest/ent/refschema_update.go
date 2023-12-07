// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/woocoos/knockout-go/integration/gentest/ent/predicate"
	"github.com/woocoos/knockout-go/integration/gentest/ent/refschema"
)

// RefSchemaUpdate is the builder for updating RefSchema entities.
type RefSchemaUpdate struct {
	config
	hooks    []Hook
	mutation *RefSchemaMutation
}

// Where appends a list predicates to the RefSchemaUpdate builder.
func (rsu *RefSchemaUpdate) Where(ps ...predicate.RefSchema) *RefSchemaUpdate {
	rsu.mutation.Where(ps...)
	return rsu
}

// SetName sets the "name" field.
func (rsu *RefSchemaUpdate) SetName(s string) *RefSchemaUpdate {
	rsu.mutation.SetName(s)
	return rsu
}

// SetNillableName sets the "name" field if the given value is not nil.
func (rsu *RefSchemaUpdate) SetNillableName(s *string) *RefSchemaUpdate {
	if s != nil {
		rsu.SetName(*s)
	}
	return rsu
}

// Mutation returns the RefSchemaMutation object of the builder.
func (rsu *RefSchemaUpdate) Mutation() *RefSchemaMutation {
	return rsu.mutation
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (rsu *RefSchemaUpdate) Save(ctx context.Context) (int, error) {
	return withHooks(ctx, rsu.sqlSave, rsu.mutation, rsu.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (rsu *RefSchemaUpdate) SaveX(ctx context.Context) int {
	affected, err := rsu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (rsu *RefSchemaUpdate) Exec(ctx context.Context) error {
	_, err := rsu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (rsu *RefSchemaUpdate) ExecX(ctx context.Context) {
	if err := rsu.Exec(ctx); err != nil {
		panic(err)
	}
}

func (rsu *RefSchemaUpdate) sqlSave(ctx context.Context) (n int, err error) {
	_spec := sqlgraph.NewUpdateSpec(refschema.Table, refschema.Columns, sqlgraph.NewFieldSpec(refschema.FieldID, field.TypeInt))
	if ps := rsu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := rsu.mutation.Name(); ok {
		_spec.SetField(refschema.FieldName, field.TypeString, value)
	}
	if n, err = sqlgraph.UpdateNodes(ctx, rsu.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{refschema.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	rsu.mutation.done = true
	return n, nil
}

// RefSchemaUpdateOne is the builder for updating a single RefSchema entity.
type RefSchemaUpdateOne struct {
	config
	fields   []string
	hooks    []Hook
	mutation *RefSchemaMutation
}

// SetName sets the "name" field.
func (rsuo *RefSchemaUpdateOne) SetName(s string) *RefSchemaUpdateOne {
	rsuo.mutation.SetName(s)
	return rsuo
}

// SetNillableName sets the "name" field if the given value is not nil.
func (rsuo *RefSchemaUpdateOne) SetNillableName(s *string) *RefSchemaUpdateOne {
	if s != nil {
		rsuo.SetName(*s)
	}
	return rsuo
}

// Mutation returns the RefSchemaMutation object of the builder.
func (rsuo *RefSchemaUpdateOne) Mutation() *RefSchemaMutation {
	return rsuo.mutation
}

// Where appends a list predicates to the RefSchemaUpdate builder.
func (rsuo *RefSchemaUpdateOne) Where(ps ...predicate.RefSchema) *RefSchemaUpdateOne {
	rsuo.mutation.Where(ps...)
	return rsuo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (rsuo *RefSchemaUpdateOne) Select(field string, fields ...string) *RefSchemaUpdateOne {
	rsuo.fields = append([]string{field}, fields...)
	return rsuo
}

// Save executes the query and returns the updated RefSchema entity.
func (rsuo *RefSchemaUpdateOne) Save(ctx context.Context) (*RefSchema, error) {
	return withHooks(ctx, rsuo.sqlSave, rsuo.mutation, rsuo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (rsuo *RefSchemaUpdateOne) SaveX(ctx context.Context) *RefSchema {
	node, err := rsuo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (rsuo *RefSchemaUpdateOne) Exec(ctx context.Context) error {
	_, err := rsuo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (rsuo *RefSchemaUpdateOne) ExecX(ctx context.Context) {
	if err := rsuo.Exec(ctx); err != nil {
		panic(err)
	}
}

func (rsuo *RefSchemaUpdateOne) sqlSave(ctx context.Context) (_node *RefSchema, err error) {
	_spec := sqlgraph.NewUpdateSpec(refschema.Table, refschema.Columns, sqlgraph.NewFieldSpec(refschema.FieldID, field.TypeInt))
	id, ok := rsuo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "RefSchema.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := rsuo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, refschema.FieldID)
		for _, f := range fields {
			if !refschema.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != refschema.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := rsuo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := rsuo.mutation.Name(); ok {
		_spec.SetField(refschema.FieldName, field.TypeString, value)
	}
	_node = &RefSchema{config: rsuo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, rsuo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{refschema.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	rsuo.mutation.done = true
	return _node, nil
}
