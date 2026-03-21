package db

import (
	"context"
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Engine struct {
	conn *sql.DB
}

func (e *Engine) Conn() *sql.DB {
	return e.conn
}

type EngineOpt func(*Engine) error

func NewEngine(ctx context.Context, dsn string, opts ...EngineOpt) (*Engine, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	engine := &Engine{
		conn: db,
	}

	for _, opt := range opts {
		if err := opt(engine); err != nil {
			return nil, err
		}
	}

	return engine, nil
}

func WithPing(ctx context.Context) EngineOpt {
	return func(e *Engine) error {
		if err := e.conn.PingContext(ctx); err != nil {
			return err
		}
		return nil
	}
}

func WithMigrate(ctx context.Context) EngineOpt {
	return func(e *Engine) error {
		_, err := e.conn.ExecContext(ctx, ddlSchema)
		return err
	}
}
