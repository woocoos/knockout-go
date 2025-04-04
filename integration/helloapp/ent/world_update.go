// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/woocoos/knockout-go/integration/helloapp/ent/predicate"
	"github.com/woocoos/knockout-go/integration/helloapp/ent/world"
)

// WorldUpdate is the builder for updating World entities.
type WorldUpdate struct {
	config
	hooks    []Hook
	mutation *WorldMutation
}

// Where appends a list predicates to the WorldUpdate builder.
func (wu *WorldUpdate) Where(ps ...predicate.World) *WorldUpdate {
	wu.mutation.Where(ps...)
	return wu
}

// SetUpdatedBy sets the "updated_by" field.
func (wu *WorldUpdate) SetUpdatedBy(i int) *WorldUpdate {
	wu.mutation.ResetUpdatedBy()
	wu.mutation.SetUpdatedBy(i)
	return wu
}

// SetNillableUpdatedBy sets the "updated_by" field if the given value is not nil.
func (wu *WorldUpdate) SetNillableUpdatedBy(i *int) *WorldUpdate {
	if i != nil {
		wu.SetUpdatedBy(*i)
	}
	return wu
}

// AddUpdatedBy adds i to the "updated_by" field.
func (wu *WorldUpdate) AddUpdatedBy(i int) *WorldUpdate {
	wu.mutation.AddUpdatedBy(i)
	return wu
}

// ClearUpdatedBy clears the value of the "updated_by" field.
func (wu *WorldUpdate) ClearUpdatedBy() *WorldUpdate {
	wu.mutation.ClearUpdatedBy()
	return wu
}

// SetUpdatedAt sets the "updated_at" field.
func (wu *WorldUpdate) SetUpdatedAt(t time.Time) *WorldUpdate {
	wu.mutation.SetUpdatedAt(t)
	return wu
}

// SetNillableUpdatedAt sets the "updated_at" field if the given value is not nil.
func (wu *WorldUpdate) SetNillableUpdatedAt(t *time.Time) *WorldUpdate {
	if t != nil {
		wu.SetUpdatedAt(*t)
	}
	return wu
}

// ClearUpdatedAt clears the value of the "updated_at" field.
func (wu *WorldUpdate) ClearUpdatedAt() *WorldUpdate {
	wu.mutation.ClearUpdatedAt()
	return wu
}

// SetDeletedAt sets the "deleted_at" field.
func (wu *WorldUpdate) SetDeletedAt(t time.Time) *WorldUpdate {
	wu.mutation.SetDeletedAt(t)
	return wu
}

// SetNillableDeletedAt sets the "deleted_at" field if the given value is not nil.
func (wu *WorldUpdate) SetNillableDeletedAt(t *time.Time) *WorldUpdate {
	if t != nil {
		wu.SetDeletedAt(*t)
	}
	return wu
}

// ClearDeletedAt clears the value of the "deleted_at" field.
func (wu *WorldUpdate) ClearDeletedAt() *WorldUpdate {
	wu.mutation.ClearDeletedAt()
	return wu
}

// SetName sets the "name" field.
func (wu *WorldUpdate) SetName(s string) *WorldUpdate {
	wu.mutation.SetName(s)
	return wu
}

// SetNillableName sets the "name" field if the given value is not nil.
func (wu *WorldUpdate) SetNillableName(s *string) *WorldUpdate {
	if s != nil {
		wu.SetName(*s)
	}
	return wu
}

// SetPowerBy sets the "power_by" field.
func (wu *WorldUpdate) SetPowerBy(s string) *WorldUpdate {
	wu.mutation.SetPowerBy(s)
	return wu
}

// SetNillablePowerBy sets the "power_by" field if the given value is not nil.
func (wu *WorldUpdate) SetNillablePowerBy(s *string) *WorldUpdate {
	if s != nil {
		wu.SetPowerBy(*s)
	}
	return wu
}

// ClearPowerBy clears the value of the "power_by" field.
func (wu *WorldUpdate) ClearPowerBy() *WorldUpdate {
	wu.mutation.ClearPowerBy()
	return wu
}

// Mutation returns the WorldMutation object of the builder.
func (wu *WorldUpdate) Mutation() *WorldMutation {
	return wu.mutation
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (wu *WorldUpdate) Save(ctx context.Context) (int, error) {
	return withHooks(ctx, wu.sqlSave, wu.mutation, wu.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (wu *WorldUpdate) SaveX(ctx context.Context) int {
	affected, err := wu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (wu *WorldUpdate) Exec(ctx context.Context) error {
	_, err := wu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (wu *WorldUpdate) ExecX(ctx context.Context) {
	if err := wu.Exec(ctx); err != nil {
		panic(err)
	}
}

func (wu *WorldUpdate) sqlSave(ctx context.Context) (n int, err error) {
	_spec := sqlgraph.NewUpdateSpec(world.Table, world.Columns, sqlgraph.NewFieldSpec(world.FieldID, field.TypeInt))
	if ps := wu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := wu.mutation.UpdatedBy(); ok {
		_spec.SetField(world.FieldUpdatedBy, field.TypeInt, value)
	}
	if value, ok := wu.mutation.AddedUpdatedBy(); ok {
		_spec.AddField(world.FieldUpdatedBy, field.TypeInt, value)
	}
	if wu.mutation.UpdatedByCleared() {
		_spec.ClearField(world.FieldUpdatedBy, field.TypeInt)
	}
	if value, ok := wu.mutation.UpdatedAt(); ok {
		_spec.SetField(world.FieldUpdatedAt, field.TypeTime, value)
	}
	if wu.mutation.UpdatedAtCleared() {
		_spec.ClearField(world.FieldUpdatedAt, field.TypeTime)
	}
	if value, ok := wu.mutation.DeletedAt(); ok {
		_spec.SetField(world.FieldDeletedAt, field.TypeTime, value)
	}
	if wu.mutation.DeletedAtCleared() {
		_spec.ClearField(world.FieldDeletedAt, field.TypeTime)
	}
	if value, ok := wu.mutation.Name(); ok {
		_spec.SetField(world.FieldName, field.TypeString, value)
	}
	if value, ok := wu.mutation.PowerBy(); ok {
		_spec.SetField(world.FieldPowerBy, field.TypeString, value)
	}
	if wu.mutation.PowerByCleared() {
		_spec.ClearField(world.FieldPowerBy, field.TypeString)
	}
	if n, err = sqlgraph.UpdateNodes(ctx, wu.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{world.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	wu.mutation.done = true
	return n, nil
}

// WorldUpdateOne is the builder for updating a single World entity.
type WorldUpdateOne struct {
	config
	fields   []string
	hooks    []Hook
	mutation *WorldMutation
}

// SetUpdatedBy sets the "updated_by" field.
func (wuo *WorldUpdateOne) SetUpdatedBy(i int) *WorldUpdateOne {
	wuo.mutation.ResetUpdatedBy()
	wuo.mutation.SetUpdatedBy(i)
	return wuo
}

// SetNillableUpdatedBy sets the "updated_by" field if the given value is not nil.
func (wuo *WorldUpdateOne) SetNillableUpdatedBy(i *int) *WorldUpdateOne {
	if i != nil {
		wuo.SetUpdatedBy(*i)
	}
	return wuo
}

// AddUpdatedBy adds i to the "updated_by" field.
func (wuo *WorldUpdateOne) AddUpdatedBy(i int) *WorldUpdateOne {
	wuo.mutation.AddUpdatedBy(i)
	return wuo
}

// ClearUpdatedBy clears the value of the "updated_by" field.
func (wuo *WorldUpdateOne) ClearUpdatedBy() *WorldUpdateOne {
	wuo.mutation.ClearUpdatedBy()
	return wuo
}

// SetUpdatedAt sets the "updated_at" field.
func (wuo *WorldUpdateOne) SetUpdatedAt(t time.Time) *WorldUpdateOne {
	wuo.mutation.SetUpdatedAt(t)
	return wuo
}

// SetNillableUpdatedAt sets the "updated_at" field if the given value is not nil.
func (wuo *WorldUpdateOne) SetNillableUpdatedAt(t *time.Time) *WorldUpdateOne {
	if t != nil {
		wuo.SetUpdatedAt(*t)
	}
	return wuo
}

// ClearUpdatedAt clears the value of the "updated_at" field.
func (wuo *WorldUpdateOne) ClearUpdatedAt() *WorldUpdateOne {
	wuo.mutation.ClearUpdatedAt()
	return wuo
}

// SetDeletedAt sets the "deleted_at" field.
func (wuo *WorldUpdateOne) SetDeletedAt(t time.Time) *WorldUpdateOne {
	wuo.mutation.SetDeletedAt(t)
	return wuo
}

// SetNillableDeletedAt sets the "deleted_at" field if the given value is not nil.
func (wuo *WorldUpdateOne) SetNillableDeletedAt(t *time.Time) *WorldUpdateOne {
	if t != nil {
		wuo.SetDeletedAt(*t)
	}
	return wuo
}

// ClearDeletedAt clears the value of the "deleted_at" field.
func (wuo *WorldUpdateOne) ClearDeletedAt() *WorldUpdateOne {
	wuo.mutation.ClearDeletedAt()
	return wuo
}

// SetName sets the "name" field.
func (wuo *WorldUpdateOne) SetName(s string) *WorldUpdateOne {
	wuo.mutation.SetName(s)
	return wuo
}

// SetNillableName sets the "name" field if the given value is not nil.
func (wuo *WorldUpdateOne) SetNillableName(s *string) *WorldUpdateOne {
	if s != nil {
		wuo.SetName(*s)
	}
	return wuo
}

// SetPowerBy sets the "power_by" field.
func (wuo *WorldUpdateOne) SetPowerBy(s string) *WorldUpdateOne {
	wuo.mutation.SetPowerBy(s)
	return wuo
}

// SetNillablePowerBy sets the "power_by" field if the given value is not nil.
func (wuo *WorldUpdateOne) SetNillablePowerBy(s *string) *WorldUpdateOne {
	if s != nil {
		wuo.SetPowerBy(*s)
	}
	return wuo
}

// ClearPowerBy clears the value of the "power_by" field.
func (wuo *WorldUpdateOne) ClearPowerBy() *WorldUpdateOne {
	wuo.mutation.ClearPowerBy()
	return wuo
}

// Mutation returns the WorldMutation object of the builder.
func (wuo *WorldUpdateOne) Mutation() *WorldMutation {
	return wuo.mutation
}

// Where appends a list predicates to the WorldUpdate builder.
func (wuo *WorldUpdateOne) Where(ps ...predicate.World) *WorldUpdateOne {
	wuo.mutation.Where(ps...)
	return wuo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (wuo *WorldUpdateOne) Select(field string, fields ...string) *WorldUpdateOne {
	wuo.fields = append([]string{field}, fields...)
	return wuo
}

// Save executes the query and returns the updated World entity.
func (wuo *WorldUpdateOne) Save(ctx context.Context) (*World, error) {
	return withHooks(ctx, wuo.sqlSave, wuo.mutation, wuo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (wuo *WorldUpdateOne) SaveX(ctx context.Context) *World {
	node, err := wuo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (wuo *WorldUpdateOne) Exec(ctx context.Context) error {
	_, err := wuo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (wuo *WorldUpdateOne) ExecX(ctx context.Context) {
	if err := wuo.Exec(ctx); err != nil {
		panic(err)
	}
}

func (wuo *WorldUpdateOne) sqlSave(ctx context.Context) (_node *World, err error) {
	_spec := sqlgraph.NewUpdateSpec(world.Table, world.Columns, sqlgraph.NewFieldSpec(world.FieldID, field.TypeInt))
	id, ok := wuo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "World.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := wuo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, world.FieldID)
		for _, f := range fields {
			if !world.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != world.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := wuo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := wuo.mutation.UpdatedBy(); ok {
		_spec.SetField(world.FieldUpdatedBy, field.TypeInt, value)
	}
	if value, ok := wuo.mutation.AddedUpdatedBy(); ok {
		_spec.AddField(world.FieldUpdatedBy, field.TypeInt, value)
	}
	if wuo.mutation.UpdatedByCleared() {
		_spec.ClearField(world.FieldUpdatedBy, field.TypeInt)
	}
	if value, ok := wuo.mutation.UpdatedAt(); ok {
		_spec.SetField(world.FieldUpdatedAt, field.TypeTime, value)
	}
	if wuo.mutation.UpdatedAtCleared() {
		_spec.ClearField(world.FieldUpdatedAt, field.TypeTime)
	}
	if value, ok := wuo.mutation.DeletedAt(); ok {
		_spec.SetField(world.FieldDeletedAt, field.TypeTime, value)
	}
	if wuo.mutation.DeletedAtCleared() {
		_spec.ClearField(world.FieldDeletedAt, field.TypeTime)
	}
	if value, ok := wuo.mutation.Name(); ok {
		_spec.SetField(world.FieldName, field.TypeString, value)
	}
	if value, ok := wuo.mutation.PowerBy(); ok {
		_spec.SetField(world.FieldPowerBy, field.TypeString, value)
	}
	if wuo.mutation.PowerByCleared() {
		_spec.ClearField(world.FieldPowerBy, field.TypeString)
	}
	_node = &World{config: wuo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, wuo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{world.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	wuo.mutation.done = true
	return _node, nil
}
