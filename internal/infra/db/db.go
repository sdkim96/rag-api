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
