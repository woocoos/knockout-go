package gentest

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/woocoos/knockout-go/codegen/entx"
	"github.com/woocoos/knockout-go/integration/gentest/ent"

	_ "github.com/mattn/go-sqlite3"
)

func TestSkipMigration(t *testing.T) {
	client, err := ent.Open("sqlite3", "file:ent?mode=memory&_fk=1")
	assert.NoError(t, err)

	err = client.Schema.Create(context.Background(), entx.SkipTablesDiffHook("ref_table"))
	assert.NoError(t, err)

	_, err = client.RefSchema.Query().All(context.Background())
	assert.Error(t, err)
	assert.ErrorContains(t, err, "no such table: ref_table")
}

func TestExcludeNodeQuery(t *testing.T) {
	gid, err := ent.GlobalID("ExGIDSchema", "1")
	assert.NoError(t, err)
	_, err = ent.FromGlobalID(gid)
	assert.ErrorContains(t, err, "invalid global identifier")
}
