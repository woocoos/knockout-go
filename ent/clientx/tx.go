package clientx

import (
	"context"
	"fmt"
)

type Tx func(ctx context.Context) (Transactor, error)

// Transactor is the interface that wraps the Commit and Rollback methods.
type Transactor interface {
	Commit() error
	Rollback() error
}

// WithTx runs the given function in a transaction.If the function returns an error, the transaction is rolled back.
//
// Example:
//
//	  WithTx(ctx, func(ctx context.Context) (clientx.Transactor, error) {
//	      return db.Tx(ctx) // db is *ent.Client
//	  }, func(itx clientx.Transactor) error {
//		     tx := itx.(*ent.Tx)
//	      // do something with tx
//	  })
func WithTx(ctx context.Context, initTxFn Tx, fn func(itx Transactor) error) error {
	tx, err := initTxFn(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()
	if err := fn(tx); err != nil {
		if rerr := tx.Rollback(); rerr != nil {
			err = fmt.Errorf("%w: rolling back transaction: %v", err, rerr)
		}
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}
	return nil
}
