package helloapp

import (
	"context"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tsingsun/woocoo/pkg/conf"
	"github.com/tsingsun/woocoo/pkg/store/sqlx"
	"github.com/woocoos/knockout-go/ent/clientx"
	"github.com/woocoos/knockout-go/integration/helloapp/ent"
	"github.com/woocoos/knockout-go/integration/helloapp/ent/enttest"
	"github.com/woocoos/knockout-go/integration/helloapp/ent/hello"
	"github.com/woocoos/knockout-go/integration/helloapp/ent/migrate"
	"github.com/woocoos/knockout-go/pkg/identity"
	"testing"
)

func Test_WithTx(t *testing.T) {
	cli := enttest.Open(t,
		"sqlite3",
		"file:withtx?mode=memory&cache=shared&_fk=1",
		enttest.WithMigrateOptions(migrate.WithForeignKeys(false)))
	defer cli.Close()
	expectedErr := errors.New("will rollback")
	t.Run("rollback", func(t *testing.T) {
		ctx := identity.WithTenantID(context.Background(), 1)
		err := clientx.WithTx(context.Background(), func(ctx context.Context) (clientx.Transactor, error) {
			return cli.Tx(ctx)
		}, func(itx clientx.Transactor) error {
			tx := itx.(*ent.Tx)
			_, err := tx.Hello.Create().SetID(1).SetName("rollback").SetTenantID(1).Save(ctx)
			require.NoError(t, err)

			return expectedErr
		})
		assert.ErrorIs(t, err, expectedErr)
		assert.Equal(t, 0, len(cli.Hello.Query().Where(hello.ID(1)).AllX(ctx)))
	})
	t.Run("commit", func(t *testing.T) {
		ctx := identity.WithTenantID(context.Background(), 1)
		err := clientx.WithTx(context.Background(), func(ctx context.Context) (clientx.Transactor, error) {
			return cli.Tx(ctx)
		}, func(itx clientx.Transactor) error {
			tx := itx.(*ent.Tx)
			_, err := tx.Hello.Create().SetID(2).SetName("rollback").SetTenantID(1).Save(ctx)
			require.NoError(t, err)

			return nil
		})
		require.NoError(t, err)
		assert.Equal(t, 1, len(cli.Hello.Query().Where(hello.ID(2)).AllX(ctx)))
	})
	t.Run("panic", func(t *testing.T) {
		ctx := identity.WithTenantID(context.Background(), 1)
		assert.Panics(t, func() {
			_ = clientx.WithTx(context.Background(), func(ctx context.Context) (clientx.Transactor, error) {
				return cli.Tx(ctx)
			}, func(itx clientx.Transactor) error {
				tx := itx.(*ent.Tx)
				_, err := tx.Hello.Create().SetID(3).SetName("rollback").SetTenantID(1).Save(ctx)
				require.NoError(t, err)
				panic(expectedErr)
			})
		})
		assert.Equal(t, 0, len(cli.Hello.Query().Where(hello.ID(3)).AllX(ctx)))
	})
}

func TestMultiInstances(t *testing.T) {
	pcfg := conf.NewFromStringMap(map[string]any{
		"store": map[string]any{
			"portal": map[string]any{
				"driverName": "sqlite3",
				"dsn":        "file:one?mode=memory&cache=shared&_fk=1",
				"multiInstances": map[string]any{
					"1": map[string]any{
						"driverName": "sqlite3",
						"dsn":        "file:two?mode=memory&cache=shared&_fk=1",
					},
				},
			},
		},
	}).Sub("store.portal")
	var pd dialect.Driver
	if pcfg.IsSet("multiInstances") {
		pd = clientx.NewRouteDriver(pcfg)
	} else {
		pd = sql.OpenDB(pcfg.String("driverName"), sqlx.NewSqlDB(pcfg))
	}
	client := ent.NewClient(ent.Driver(pd))
	ctx := identity.WithTenantID(context.Background(), 1)
	err := client.Schema.Create(ctx)
	assert.NoError(t, err)
}
