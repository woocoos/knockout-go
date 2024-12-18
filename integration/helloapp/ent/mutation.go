// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/woocoos/knockout-go/integration/helloapp/ent/hello"
	"github.com/woocoos/knockout-go/integration/helloapp/ent/predicate"
	"github.com/woocoos/knockout-go/integration/helloapp/ent/world"
)

const (
	// Operation types.
	OpCreate    = ent.OpCreate
	OpDelete    = ent.OpDelete
	OpDeleteOne = ent.OpDeleteOne
	OpUpdate    = ent.OpUpdate
	OpUpdateOne = ent.OpUpdateOne

	// Node types.
	TypeHello = "Hello"
	TypeWorld = "World"
)

// HelloMutation represents an operation that mutates the Hello nodes in the graph.
type HelloMutation struct {
	config
	op            Op
	typ           string
	id            *int
	name          *string
	tenant_id     *int
	addtenant_id  *int
	clearedFields map[string]struct{}
	done          bool
	oldValue      func(context.Context) (*Hello, error)
	predicates    []predicate.Hello
}

var _ ent.Mutation = (*HelloMutation)(nil)

// helloOption allows management of the mutation configuration using functional options.
type helloOption func(*HelloMutation)

// newHelloMutation creates new mutation for the Hello entity.
func newHelloMutation(c config, op Op, opts ...helloOption) *HelloMutation {
	m := &HelloMutation{
		config:        c,
		op:            op,
		typ:           TypeHello,
		clearedFields: make(map[string]struct{}),
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// withHelloID sets the ID field of the mutation.
func withHelloID(id int) helloOption {
	return func(m *HelloMutation) {
		var (
			err   error
			once  sync.Once
			value *Hello
		)
		m.oldValue = func(ctx context.Context) (*Hello, error) {
			once.Do(func() {
				if m.done {
					err = errors.New("querying old values post mutation is not allowed")
				} else {
					value, err = m.Client().Hello.Get(ctx, id)
				}
			})
			return value, err
		}
		m.id = &id
	}
}

// withHello sets the old Hello of the mutation.
func withHello(node *Hello) helloOption {
	return func(m *HelloMutation) {
		m.oldValue = func(context.Context) (*Hello, error) {
			return node, nil
		}
		m.id = &node.ID
	}
}

// Client returns a new `ent.Client` from the mutation. If the mutation was
// executed in a transaction (ent.Tx), a transactional client is returned.
func (m HelloMutation) Client() *Client {
	client := &Client{config: m.config}
	client.init()
	return client
}

// Tx returns an `ent.Tx` for mutations that were executed in transactions;
// it returns an error otherwise.
func (m HelloMutation) Tx() (*Tx, error) {
	if _, ok := m.driver.(*txDriver); !ok {
		return nil, errors.New("ent: mutation is not running in a transaction")
	}
	tx := &Tx{config: m.config}
	tx.init()
	return tx, nil
}

// SetID sets the value of the id field. Note that this
// operation is only accepted on creation of Hello entities.
func (m *HelloMutation) SetID(id int) {
	m.id = &id
}

// ID returns the ID value in the mutation. Note that the ID is only available
// if it was provided to the builder or after it was returned from the database.
func (m *HelloMutation) ID() (id int, exists bool) {
	if m.id == nil {
		return
	}
	return *m.id, true
}

// IDs queries the database and returns the entity ids that match the mutation's predicate.
// That means, if the mutation is applied within a transaction with an isolation level such
// as sql.LevelSerializable, the returned ids match the ids of the rows that will be updated
// or updated by the mutation.
func (m *HelloMutation) IDs(ctx context.Context) ([]int, error) {
	switch {
	case m.op.Is(OpUpdateOne | OpDeleteOne):
		id, exists := m.ID()
		if exists {
			return []int{id}, nil
		}
		fallthrough
	case m.op.Is(OpUpdate | OpDelete):
		return m.Client().Hello.Query().Where(m.predicates...).IDs(ctx)
	default:
		return nil, fmt.Errorf("IDs is not allowed on %s operations", m.op)
	}
}

// SetName sets the "name" field.
func (m *HelloMutation) SetName(s string) {
	m.name = &s
}

// Name returns the value of the "name" field in the mutation.
func (m *HelloMutation) Name() (r string, exists bool) {
	v := m.name
	if v == nil {
		return
	}
	return *v, true
}

// OldName returns the old "name" field's value of the Hello entity.
// If the Hello object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *HelloMutation) OldName(ctx context.Context) (v string, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, errors.New("OldName is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, errors.New("OldName requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldName: %w", err)
	}
	return oldValue.Name, nil
}

// ResetName resets all changes to the "name" field.
func (m *HelloMutation) ResetName() {
	m.name = nil
}

// SetTenantID sets the "tenant_id" field.
func (m *HelloMutation) SetTenantID(i int) {
	m.tenant_id = &i
	m.addtenant_id = nil
}

// TenantID returns the value of the "tenant_id" field in the mutation.
func (m *HelloMutation) TenantID() (r int, exists bool) {
	v := m.tenant_id
	if v == nil {
		return
	}
	return *v, true
}

// OldTenantID returns the old "tenant_id" field's value of the Hello entity.
// If the Hello object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *HelloMutation) OldTenantID(ctx context.Context) (v int, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, errors.New("OldTenantID is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, errors.New("OldTenantID requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldTenantID: %w", err)
	}
	return oldValue.TenantID, nil
}

// AddTenantID adds i to the "tenant_id" field.
func (m *HelloMutation) AddTenantID(i int) {
	if m.addtenant_id != nil {
		*m.addtenant_id += i
	} else {
		m.addtenant_id = &i
	}
}

// AddedTenantID returns the value that was added to the "tenant_id" field in this mutation.
func (m *HelloMutation) AddedTenantID() (r int, exists bool) {
	v := m.addtenant_id
	if v == nil {
		return
	}
	return *v, true
}

// ResetTenantID resets all changes to the "tenant_id" field.
func (m *HelloMutation) ResetTenantID() {
	m.tenant_id = nil
	m.addtenant_id = nil
}

// Where appends a list predicates to the HelloMutation builder.
func (m *HelloMutation) Where(ps ...predicate.Hello) {
	m.predicates = append(m.predicates, ps...)
}

// WhereP appends storage-level predicates to the HelloMutation builder. Using this method,
// users can use type-assertion to append predicates that do not depend on any generated package.
func (m *HelloMutation) WhereP(ps ...func(*sql.Selector)) {
	p := make([]predicate.Hello, len(ps))
	for i := range ps {
		p[i] = ps[i]
	}
	m.Where(p...)
}

// Op returns the operation name.
func (m *HelloMutation) Op() Op {
	return m.op
}

// SetOp allows setting the mutation operation.
func (m *HelloMutation) SetOp(op Op) {
	m.op = op
}

// Type returns the node type of this mutation (Hello).
func (m *HelloMutation) Type() string {
	return m.typ
}

// Fields returns all fields that were changed during this mutation. Note that in
// order to get all numeric fields that were incremented/decremented, call
// AddedFields().
func (m *HelloMutation) Fields() []string {
	fields := make([]string, 0, 2)
	if m.name != nil {
		fields = append(fields, hello.FieldName)
	}
	if m.tenant_id != nil {
		fields = append(fields, hello.FieldTenantID)
	}
	return fields
}

// Field returns the value of a field with the given name. The second boolean
// return value indicates that this field was not set, or was not defined in the
// schema.
func (m *HelloMutation) Field(name string) (ent.Value, bool) {
	switch name {
	case hello.FieldName:
		return m.Name()
	case hello.FieldTenantID:
		return m.TenantID()
	}
	return nil, false
}

// OldField returns the old value of the field from the database. An error is
// returned if the mutation operation is not UpdateOne, or the query to the
// database failed.
func (m *HelloMutation) OldField(ctx context.Context, name string) (ent.Value, error) {
	switch name {
	case hello.FieldName:
		return m.OldName(ctx)
	case hello.FieldTenantID:
		return m.OldTenantID(ctx)
	}
	return nil, fmt.Errorf("unknown Hello field %s", name)
}

// SetField sets the value of a field with the given name. It returns an error if
// the field is not defined in the schema, or if the type mismatched the field
// type.
func (m *HelloMutation) SetField(name string, value ent.Value) error {
	switch name {
	case hello.FieldName:
		v, ok := value.(string)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetName(v)
		return nil
	case hello.FieldTenantID:
		v, ok := value.(int)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetTenantID(v)
		return nil
	}
	return fmt.Errorf("unknown Hello field %s", name)
}

// AddedFields returns all numeric fields that were incremented/decremented during
// this mutation.
func (m *HelloMutation) AddedFields() []string {
	var fields []string
	if m.addtenant_id != nil {
		fields = append(fields, hello.FieldTenantID)
	}
	return fields
}

// AddedField returns the numeric value that was incremented/decremented on a field
// with the given name. The second boolean return value indicates that this field
// was not set, or was not defined in the schema.
func (m *HelloMutation) AddedField(name string) (ent.Value, bool) {
	switch name {
	case hello.FieldTenantID:
		return m.AddedTenantID()
	}
	return nil, false
}

// AddField adds the value to the field with the given name. It returns an error if
// the field is not defined in the schema, or if the type mismatched the field
// type.
func (m *HelloMutation) AddField(name string, value ent.Value) error {
	switch name {
	case hello.FieldTenantID:
		v, ok := value.(int)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.AddTenantID(v)
		return nil
	}
	return fmt.Errorf("unknown Hello numeric field %s", name)
}

// ClearedFields returns all nullable fields that were cleared during this
// mutation.
func (m *HelloMutation) ClearedFields() []string {
	return nil
}

// FieldCleared returns a boolean indicating if a field with the given name was
// cleared in this mutation.
func (m *HelloMutation) FieldCleared(name string) bool {
	_, ok := m.clearedFields[name]
	return ok
}

// ClearField clears the value of the field with the given name. It returns an
// error if the field is not defined in the schema.
func (m *HelloMutation) ClearField(name string) error {
	return fmt.Errorf("unknown Hello nullable field %s", name)
}

// ResetField resets all changes in the mutation for the field with the given name.
// It returns an error if the field is not defined in the schema.
func (m *HelloMutation) ResetField(name string) error {
	switch name {
	case hello.FieldName:
		m.ResetName()
		return nil
	case hello.FieldTenantID:
		m.ResetTenantID()
		return nil
	}
	return fmt.Errorf("unknown Hello field %s", name)
}

// AddedEdges returns all edge names that were set/added in this mutation.
func (m *HelloMutation) AddedEdges() []string {
	edges := make([]string, 0, 0)
	return edges
}

// AddedIDs returns all IDs (to other nodes) that were added for the given edge
// name in this mutation.
func (m *HelloMutation) AddedIDs(name string) []ent.Value {
	return nil
}

// RemovedEdges returns all edge names that were removed in this mutation.
func (m *HelloMutation) RemovedEdges() []string {
	edges := make([]string, 0, 0)
	return edges
}

// RemovedIDs returns all IDs (to other nodes) that were removed for the edge with
// the given name in this mutation.
func (m *HelloMutation) RemovedIDs(name string) []ent.Value {
	return nil
}

// ClearedEdges returns all edge names that were cleared in this mutation.
func (m *HelloMutation) ClearedEdges() []string {
	edges := make([]string, 0, 0)
	return edges
}

// EdgeCleared returns a boolean which indicates if the edge with the given name
// was cleared in this mutation.
func (m *HelloMutation) EdgeCleared(name string) bool {
	return false
}

// ClearEdge clears the value of the edge with the given name. It returns an error
// if that edge is not defined in the schema.
func (m *HelloMutation) ClearEdge(name string) error {
	return fmt.Errorf("unknown Hello unique edge %s", name)
}

// ResetEdge resets all changes to the edge with the given name in this mutation.
// It returns an error if the edge is not defined in the schema.
func (m *HelloMutation) ResetEdge(name string) error {
	return fmt.Errorf("unknown Hello edge %s", name)
}

// WorldMutation represents an operation that mutates the World nodes in the graph.
type WorldMutation struct {
	config
	op            Op
	typ           string
	id            *int
	deleted_at    *time.Time
	tenant_id     *int
	addtenant_id  *int
	name          *string
	power_by      *string
	clearedFields map[string]struct{}
	done          bool
	oldValue      func(context.Context) (*World, error)
	predicates    []predicate.World
}

var _ ent.Mutation = (*WorldMutation)(nil)

// worldOption allows management of the mutation configuration using functional options.
type worldOption func(*WorldMutation)

// newWorldMutation creates new mutation for the World entity.
func newWorldMutation(c config, op Op, opts ...worldOption) *WorldMutation {
	m := &WorldMutation{
		config:        c,
		op:            op,
		typ:           TypeWorld,
		clearedFields: make(map[string]struct{}),
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// withWorldID sets the ID field of the mutation.
func withWorldID(id int) worldOption {
	return func(m *WorldMutation) {
		var (
			err   error
			once  sync.Once
			value *World
		)
		m.oldValue = func(ctx context.Context) (*World, error) {
			once.Do(func() {
				if m.done {
					err = errors.New("querying old values post mutation is not allowed")
				} else {
					value, err = m.Client().World.Get(ctx, id)
				}
			})
			return value, err
		}
		m.id = &id
	}
}

// withWorld sets the old World of the mutation.
func withWorld(node *World) worldOption {
	return func(m *WorldMutation) {
		m.oldValue = func(context.Context) (*World, error) {
			return node, nil
		}
		m.id = &node.ID
	}
}

// Client returns a new `ent.Client` from the mutation. If the mutation was
// executed in a transaction (ent.Tx), a transactional client is returned.
func (m WorldMutation) Client() *Client {
	client := &Client{config: m.config}
	client.init()
	return client
}

// Tx returns an `ent.Tx` for mutations that were executed in transactions;
// it returns an error otherwise.
func (m WorldMutation) Tx() (*Tx, error) {
	if _, ok := m.driver.(*txDriver); !ok {
		return nil, errors.New("ent: mutation is not running in a transaction")
	}
	tx := &Tx{config: m.config}
	tx.init()
	return tx, nil
}

// SetID sets the value of the id field. Note that this
// operation is only accepted on creation of World entities.
func (m *WorldMutation) SetID(id int) {
	m.id = &id
}

// ID returns the ID value in the mutation. Note that the ID is only available
// if it was provided to the builder or after it was returned from the database.
func (m *WorldMutation) ID() (id int, exists bool) {
	if m.id == nil {
		return
	}
	return *m.id, true
}

// IDs queries the database and returns the entity ids that match the mutation's predicate.
// That means, if the mutation is applied within a transaction with an isolation level such
// as sql.LevelSerializable, the returned ids match the ids of the rows that will be updated
// or updated by the mutation.
func (m *WorldMutation) IDs(ctx context.Context) ([]int, error) {
	switch {
	case m.op.Is(OpUpdateOne | OpDeleteOne):
		id, exists := m.ID()
		if exists {
			return []int{id}, nil
		}
		fallthrough
	case m.op.Is(OpUpdate | OpDelete):
		return m.Client().World.Query().Where(m.predicates...).IDs(ctx)
	default:
		return nil, fmt.Errorf("IDs is not allowed on %s operations", m.op)
	}
}

// SetDeletedAt sets the "deleted_at" field.
func (m *WorldMutation) SetDeletedAt(t time.Time) {
	m.deleted_at = &t
}

// DeletedAt returns the value of the "deleted_at" field in the mutation.
func (m *WorldMutation) DeletedAt() (r time.Time, exists bool) {
	v := m.deleted_at
	if v == nil {
		return
	}
	return *v, true
}

// OldDeletedAt returns the old "deleted_at" field's value of the World entity.
// If the World object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *WorldMutation) OldDeletedAt(ctx context.Context) (v time.Time, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, errors.New("OldDeletedAt is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, errors.New("OldDeletedAt requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldDeletedAt: %w", err)
	}
	return oldValue.DeletedAt, nil
}

// ClearDeletedAt clears the value of the "deleted_at" field.
func (m *WorldMutation) ClearDeletedAt() {
	m.deleted_at = nil
	m.clearedFields[world.FieldDeletedAt] = struct{}{}
}

// DeletedAtCleared returns if the "deleted_at" field was cleared in this mutation.
func (m *WorldMutation) DeletedAtCleared() bool {
	_, ok := m.clearedFields[world.FieldDeletedAt]
	return ok
}

// ResetDeletedAt resets all changes to the "deleted_at" field.
func (m *WorldMutation) ResetDeletedAt() {
	m.deleted_at = nil
	delete(m.clearedFields, world.FieldDeletedAt)
}

// SetTenantID sets the "tenant_id" field.
func (m *WorldMutation) SetTenantID(i int) {
	m.tenant_id = &i
	m.addtenant_id = nil
}

// TenantID returns the value of the "tenant_id" field in the mutation.
func (m *WorldMutation) TenantID() (r int, exists bool) {
	v := m.tenant_id
	if v == nil {
		return
	}
	return *v, true
}

// OldTenantID returns the old "tenant_id" field's value of the World entity.
// If the World object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *WorldMutation) OldTenantID(ctx context.Context) (v int, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, errors.New("OldTenantID is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, errors.New("OldTenantID requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldTenantID: %w", err)
	}
	return oldValue.TenantID, nil
}

// AddTenantID adds i to the "tenant_id" field.
func (m *WorldMutation) AddTenantID(i int) {
	if m.addtenant_id != nil {
		*m.addtenant_id += i
	} else {
		m.addtenant_id = &i
	}
}

// AddedTenantID returns the value that was added to the "tenant_id" field in this mutation.
func (m *WorldMutation) AddedTenantID() (r int, exists bool) {
	v := m.addtenant_id
	if v == nil {
		return
	}
	return *v, true
}

// ResetTenantID resets all changes to the "tenant_id" field.
func (m *WorldMutation) ResetTenantID() {
	m.tenant_id = nil
	m.addtenant_id = nil
}

// SetName sets the "name" field.
func (m *WorldMutation) SetName(s string) {
	m.name = &s
}

// Name returns the value of the "name" field in the mutation.
func (m *WorldMutation) Name() (r string, exists bool) {
	v := m.name
	if v == nil {
		return
	}
	return *v, true
}

// OldName returns the old "name" field's value of the World entity.
// If the World object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *WorldMutation) OldName(ctx context.Context) (v string, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, errors.New("OldName is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, errors.New("OldName requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldName: %w", err)
	}
	return oldValue.Name, nil
}

// ResetName resets all changes to the "name" field.
func (m *WorldMutation) ResetName() {
	m.name = nil
}

// SetPowerBy sets the "power_by" field.
func (m *WorldMutation) SetPowerBy(s string) {
	m.power_by = &s
}

// PowerBy returns the value of the "power_by" field in the mutation.
func (m *WorldMutation) PowerBy() (r string, exists bool) {
	v := m.power_by
	if v == nil {
		return
	}
	return *v, true
}

// OldPowerBy returns the old "power_by" field's value of the World entity.
// If the World object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *WorldMutation) OldPowerBy(ctx context.Context) (v string, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, errors.New("OldPowerBy is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, errors.New("OldPowerBy requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldPowerBy: %w", err)
	}
	return oldValue.PowerBy, nil
}

// ClearPowerBy clears the value of the "power_by" field.
func (m *WorldMutation) ClearPowerBy() {
	m.power_by = nil
	m.clearedFields[world.FieldPowerBy] = struct{}{}
}

// PowerByCleared returns if the "power_by" field was cleared in this mutation.
func (m *WorldMutation) PowerByCleared() bool {
	_, ok := m.clearedFields[world.FieldPowerBy]
	return ok
}

// ResetPowerBy resets all changes to the "power_by" field.
func (m *WorldMutation) ResetPowerBy() {
	m.power_by = nil
	delete(m.clearedFields, world.FieldPowerBy)
}

// Where appends a list predicates to the WorldMutation builder.
func (m *WorldMutation) Where(ps ...predicate.World) {
	m.predicates = append(m.predicates, ps...)
}

// WhereP appends storage-level predicates to the WorldMutation builder. Using this method,
// users can use type-assertion to append predicates that do not depend on any generated package.
func (m *WorldMutation) WhereP(ps ...func(*sql.Selector)) {
	p := make([]predicate.World, len(ps))
	for i := range ps {
		p[i] = ps[i]
	}
	m.Where(p...)
}

// Op returns the operation name.
func (m *WorldMutation) Op() Op {
	return m.op
}

// SetOp allows setting the mutation operation.
func (m *WorldMutation) SetOp(op Op) {
	m.op = op
}

// Type returns the node type of this mutation (World).
func (m *WorldMutation) Type() string {
	return m.typ
}

// Fields returns all fields that were changed during this mutation. Note that in
// order to get all numeric fields that were incremented/decremented, call
// AddedFields().
func (m *WorldMutation) Fields() []string {
	fields := make([]string, 0, 4)
	if m.deleted_at != nil {
		fields = append(fields, world.FieldDeletedAt)
	}
	if m.tenant_id != nil {
		fields = append(fields, world.FieldTenantID)
	}
	if m.name != nil {
		fields = append(fields, world.FieldName)
	}
	if m.power_by != nil {
		fields = append(fields, world.FieldPowerBy)
	}
	return fields
}

// Field returns the value of a field with the given name. The second boolean
// return value indicates that this field was not set, or was not defined in the
// schema.
func (m *WorldMutation) Field(name string) (ent.Value, bool) {
	switch name {
	case world.FieldDeletedAt:
		return m.DeletedAt()
	case world.FieldTenantID:
		return m.TenantID()
	case world.FieldName:
		return m.Name()
	case world.FieldPowerBy:
		return m.PowerBy()
	}
	return nil, false
}

// OldField returns the old value of the field from the database. An error is
// returned if the mutation operation is not UpdateOne, or the query to the
// database failed.
func (m *WorldMutation) OldField(ctx context.Context, name string) (ent.Value, error) {
	switch name {
	case world.FieldDeletedAt:
		return m.OldDeletedAt(ctx)
	case world.FieldTenantID:
		return m.OldTenantID(ctx)
	case world.FieldName:
		return m.OldName(ctx)
	case world.FieldPowerBy:
		return m.OldPowerBy(ctx)
	}
	return nil, fmt.Errorf("unknown World field %s", name)
}

// SetField sets the value of a field with the given name. It returns an error if
// the field is not defined in the schema, or if the type mismatched the field
// type.
func (m *WorldMutation) SetField(name string, value ent.Value) error {
	switch name {
	case world.FieldDeletedAt:
		v, ok := value.(time.Time)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetDeletedAt(v)
		return nil
	case world.FieldTenantID:
		v, ok := value.(int)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetTenantID(v)
		return nil
	case world.FieldName:
		v, ok := value.(string)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetName(v)
		return nil
	case world.FieldPowerBy:
		v, ok := value.(string)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetPowerBy(v)
		return nil
	}
	return fmt.Errorf("unknown World field %s", name)
}

// AddedFields returns all numeric fields that were incremented/decremented during
// this mutation.
func (m *WorldMutation) AddedFields() []string {
	var fields []string
	if m.addtenant_id != nil {
		fields = append(fields, world.FieldTenantID)
	}
	return fields
}

// AddedField returns the numeric value that was incremented/decremented on a field
// with the given name. The second boolean return value indicates that this field
// was not set, or was not defined in the schema.
func (m *WorldMutation) AddedField(name string) (ent.Value, bool) {
	switch name {
	case world.FieldTenantID:
		return m.AddedTenantID()
	}
	return nil, false
}

// AddField adds the value to the field with the given name. It returns an error if
// the field is not defined in the schema, or if the type mismatched the field
// type.
func (m *WorldMutation) AddField(name string, value ent.Value) error {
	switch name {
	case world.FieldTenantID:
		v, ok := value.(int)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.AddTenantID(v)
		return nil
	}
	return fmt.Errorf("unknown World numeric field %s", name)
}

// ClearedFields returns all nullable fields that were cleared during this
// mutation.
func (m *WorldMutation) ClearedFields() []string {
	var fields []string
	if m.FieldCleared(world.FieldDeletedAt) {
		fields = append(fields, world.FieldDeletedAt)
	}
	if m.FieldCleared(world.FieldPowerBy) {
		fields = append(fields, world.FieldPowerBy)
	}
	return fields
}

// FieldCleared returns a boolean indicating if a field with the given name was
// cleared in this mutation.
func (m *WorldMutation) FieldCleared(name string) bool {
	_, ok := m.clearedFields[name]
	return ok
}

// ClearField clears the value of the field with the given name. It returns an
// error if the field is not defined in the schema.
func (m *WorldMutation) ClearField(name string) error {
	switch name {
	case world.FieldDeletedAt:
		m.ClearDeletedAt()
		return nil
	case world.FieldPowerBy:
		m.ClearPowerBy()
		return nil
	}
	return fmt.Errorf("unknown World nullable field %s", name)
}

// ResetField resets all changes in the mutation for the field with the given name.
// It returns an error if the field is not defined in the schema.
func (m *WorldMutation) ResetField(name string) error {
	switch name {
	case world.FieldDeletedAt:
		m.ResetDeletedAt()
		return nil
	case world.FieldTenantID:
		m.ResetTenantID()
		return nil
	case world.FieldName:
		m.ResetName()
		return nil
	case world.FieldPowerBy:
		m.ResetPowerBy()
		return nil
	}
	return fmt.Errorf("unknown World field %s", name)
}

// AddedEdges returns all edge names that were set/added in this mutation.
func (m *WorldMutation) AddedEdges() []string {
	edges := make([]string, 0, 0)
	return edges
}

// AddedIDs returns all IDs (to other nodes) that were added for the given edge
// name in this mutation.
func (m *WorldMutation) AddedIDs(name string) []ent.Value {
	return nil
}

// RemovedEdges returns all edge names that were removed in this mutation.
func (m *WorldMutation) RemovedEdges() []string {
	edges := make([]string, 0, 0)
	return edges
}

// RemovedIDs returns all IDs (to other nodes) that were removed for the edge with
// the given name in this mutation.
func (m *WorldMutation) RemovedIDs(name string) []ent.Value {
	return nil
}

// ClearedEdges returns all edge names that were cleared in this mutation.
func (m *WorldMutation) ClearedEdges() []string {
	edges := make([]string, 0, 0)
	return edges
}

// EdgeCleared returns a boolean which indicates if the edge with the given name
// was cleared in this mutation.
func (m *WorldMutation) EdgeCleared(name string) bool {
	return false
}

// ClearEdge clears the value of the edge with the given name. It returns an error
// if that edge is not defined in the schema.
func (m *WorldMutation) ClearEdge(name string) error {
	return fmt.Errorf("unknown World unique edge %s", name)
}

// ResetEdge resets all changes to the edge with the given name in this mutation.
// It returns an error if the edge is not defined in the schema.
func (m *WorldMutation) ResetEdge(name string) error {
	return fmt.Errorf("unknown World edge %s", name)
}
