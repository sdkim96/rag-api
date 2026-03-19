package db

import (
	"context"
	"database/sql"
	"fmt"
)

func NewSourceNS(name string) namespace {
	return namespace{scheme: "source", name: name}
}

func NewIndexNS(name string) namespace {
	return namespace{scheme: "index", name: name}
}

func NewToolNS(name string) namespace {
	return namespace{scheme: "tool", name: name}
}

type namespace struct {
	scheme string
	name   string
}

func (n namespace) String() string {
	return fmt.Sprintf("%s://%s", n.scheme, n.name)
}

func (e *Engine) ReadAll(ctx context.Context, namespace namespace, opts ...ReadOption) ([][]byte, error) {
	options := &readOptions{}
	for _, opt := range opts {
		opt(options)
	}

	stmt := `SELECT value FROM store WHERE namespace = $1 AND deleted_at IS NULL`
	args := []any{namespace.String(), options.offset, options.limit}

	//TODO: Change to the metadata like
	if options.keyword != "" {
		stmt += ` AND value::text ILIKE '%' || $4 || '%'`
		args = append(args, options.keyword)
	}

	stmt += ` OFFSET $2 LIMIT $3`

	return readAll(ctx, e.Conn(), stmt, args...)
}

type ReadOption func(*readOptions)

type readOptions struct {
	offset  int
	limit   int
	keyword string
}

func WithOffset(offset int) ReadOption {
	return func(opts *readOptions) {
		opts.offset = offset
	}
}

func WithLimit(limit int) ReadOption {
	return func(opts *readOptions) {
		opts.limit = limit
	}
}

func WithKeyword(keyword string) ReadOption {
	return func(opts *readOptions) {
		opts.keyword = keyword
	}
}

func (e *Engine) Read(ctx context.Context, namespace namespace, key string) ([]byte, error) {
	return read(ctx, e.Conn(),
		`SELECT value FROM store 
		 WHERE namespace = $1 AND key = $2 AND deleted_at IS NULL`,
		namespace.String(), key,
	)
}

func (e *Engine) Write(ctx context.Context, namespace namespace, key string, value []byte) error {
	return write(ctx, e.Conn(),
		`INSERT INTO store (namespace, key, value)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (namespace, key) DO UPDATE SET value = $3`,
		namespace.String(), key, value,
	)
}

func (e *Engine) Delete(ctx context.Context, namespace namespace, key string) error {
	return write(ctx, e.Conn(),
		`UPDATE store SET deleted_at = NOW()
		 WHERE namespace = $1 AND key = $2`,
		namespace.String(), key,
	)
}

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
