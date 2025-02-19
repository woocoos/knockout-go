package gentest

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/woocoos/knockout-go/integration/nocache/ent"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestSkipMigration(t *testing.T) {
	client, err := ent.Open("sqlite3", "file:nocache?mode=memory&_fk=1")
	assert.NoError(t, err)

	err = client.Schema.Create(context.Background())
	assert.NoError(t, err)
	client.NoCache.Create().SetUserID(1).SetName("name").SaveX(context.Background())
	gid, err := ent.GlobalID("NoCache", "1")
	assert.NoError(t, err)
	nd, err := client.NoderEx(context.Background(), gid)
	assert.NoError(t, err)
	assert.IsType(t, &ent.NoCache{}, nd)
}
