package db

import (
	"context"
	"database/sql"

	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/pkg/utils"
)

func Exec(ctx context.Context, client Client, query string, args ...any) (sql.Result, error) {
	return client.Exec(ctx, query, args...)
}

func Query[T any](ctx context.Context, client Client, query string, args ...any) (ts []T, err error) {
	rows, err := client.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	// if T is *entity.Entity, return the rows as is
	var t T
	if _, ok := any(t).(*entity.Entity); ok {
		return any(rows).([]T), nil
	}

	// if T is entity.Entity, deference the rows
	if _, ok := any(t).(entity.Entity); ok {
		return utils.Map(rows, func(row *entity.Entity) T {
			return any(*row).(T)
		}), nil
	}

	// if T is not *entity.Entity, bind the rows to the struct T
	for _, row := range rows {
		var t T
		if err := BindStruct(row, &t); err != nil {
			return nil, err
		}

		ts = append(ts, t)
	}

	return ts, nil
}
