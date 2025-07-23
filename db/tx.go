package db

import (
	"context"
	"fmt"

	"github.com/fastschema/fastschema/pkg/errors"
)

func WithTx(client Client, c context.Context, fn func(tx Client) error) (err error) {
	var tx Client

	if tx, err = client.Tx(c); err != nil {
		return errors.BadRequest("error while starting transaction")
	}

	defer func() {
		if err != nil {
			if e := tx.Rollback(); e != nil {
				err = fmt.Errorf("error while rolling back transaction: %w, original error: %w", e, err)
			}
		} else {
			if e := tx.Commit(); e != nil {
				err = fmt.Errorf("error while committing transaction: %w", e)
			}
		}
	}()

	err = fn(tx)

	return
}
