package db

import (
	"context"
	"database/sql"
	"fmt"
)

func readAll(ctx context.Context, conn *sql.DB, query string, args ...any) ([][]byte, error) {
	rows, err := conn.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results [][]byte
	for rows.Next() {
		var value []byte
		if err := rows.Scan(&value); err != nil {
			return nil, err
		}
		results = append(results, value)
	}
	return results, rows.Err()
}

func read(ctx context.Context, conn *sql.DB, query string, args ...any) ([]byte, error) {
	var value []byte
	err := conn.QueryRowContext(ctx, query, args...).Scan(&value)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return value, nil
}

func write(ctx context.Context, conn *sql.DB, query string, args ...any) error {
	result, err := conn.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("no rows affected")
	}
	return nil
}

func search(ctx context.Context, conn *sql.DB, query string, args ...any) (*sql.Rows, error) {
	return conn.QueryContext(ctx, query, args...)
}
