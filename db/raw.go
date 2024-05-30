package db

import (
	"context"
	"database/sql"

	"github.com/fastschema/fastschema/schema"
)

func Exec(ctx context.Context, client Client, query string, args ...any) (sql.Result, error) {
	return client.Exec(ctx, query, args)
}

func RawQuery[T any](ctx context.Context, client Client, query string, args ...any) (ts []T, err error) {
	rows, err := client.Query(ctx, query, args)
	if err != nil {
		return nil, err
	}

	// if T is *schema.Entity, return the rows as is
	var t T
	if _, ok := any(t).(*schema.Entity); ok {
		return any(rows).([]T), nil
	}

	// if T is not *schema.Entity, bind the rows to the struct T
	for _, row := range rows {
		var t T
		if err := BindStruct(row, &t); err != nil {
			return nil, err
		}

		ts = append(ts, t)
	}

	return ts, nil
}
